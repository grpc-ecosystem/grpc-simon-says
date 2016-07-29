/* Copyright 2015 Google Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
==============================================================================*/

package simonsays

import (
	"errors"
	"io"
	"log"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/garyburd/redigo/redis"
	"github.com/grpc-simonsays/simonsays-server/simonsays/logger"
	"golang.org/x/net/context"
)

// SimonSays is the data structure that implements the SimonSaysServer
// interface for our gRPC server.
type SimonSays struct {
	pool *redis.Pool
}

// Version is the current version of this implementation of Simon Says.
const Version string = "v0.1e"

// NewSimonSays Create a new Simon Says.
func NewSimonSays(address string) (*SimonSays, error) {
	log.Printf("[Info][Server] Starting Server: %v", Version)

	if address == "" {
		address = ":6379"
	}

	s := &SimonSays{pool: newPool(address)}

	log.Printf("[Info][Redis] Connecting: %v", address)
	return s, s.pingRedis()
}

// Close closes all resources.
func (s SimonSays) Close() error {
	return s.pool.Close()
}

// Game function is an implementation of the gRPC Game Service.
// When connected, this is the main functionality of running a
// Game for the connected player.
func (s *SimonSays) Game(stream SimonSays_GameServer) error {
	ctx := stream.Context()
	defer logger.Clear(ctx)

	lc := "Game"
	// first let's get the player
	req, err := receiveRequest(stream)
	if err != nil {
		return err
	}
	player := req.GetJoin()
	if player == nil {
		logger.Error(ctx, lc, "Player was nil on initial join request. %v", req)
		return errors.New("Player was nil on initial join request.")
	}
	logger.Set(ctx, "Player", player.Id)
	logger.Info(ctx, lc, "Player %#v is attempting to join.", player)

	// find what game to join
	con := s.pool.Get()
	defer con.Close()
	game, isNew, err := findGame(ctx, con)

	if err != nil {
		return err
	}
	logger.Set(ctx, "Game", game.ID)
	logger.Info(ctx, lc, "Connecting to game %v. New?: %v", game.ID, isNew)

	logger.Info(ctx, lc, "Start to receive PubSub messages")

	// make sure that you always unjoin, if something happens to go wrong.
	defer func() {
		con := s.pool.Get()
		err := closeOpenGame(ctx, con, game)
		if err != nil {
			logger.Error(ctx, lc, "Error attempting to close game. %v", err)
		}
		err = con.Close()
		if err != nil {
			logger.Error(ctx, lc, "Error closing close open game connection. %v", err)
		}
	}()

	// make sure that at the end, you always unsubscribe.
	defer func() {
		con := s.pool.Get()
		err := redis.PubSubConn{con}.Unsubscribe(game.ID)
		if err != nil {
			logger.Error(ctx, lc, "Error unsubscribing from Game Topic %v, %v", game.ID, err)
		}
		err = con.Close()
		if err != nil {
			logger.Error(ctx, lc, "Error closing unsubscribe connection. %v", err)
		}

	}()

	msgs, err := subscribe(ctx, s.pool.Get(), game)

	if err != nil {
		return err
	}

	if err := connectGame(ctx, con, game, player, isNew); err != nil {
		return err
	}

	// subscribe to incoming key events, and get back a channel of errors.
	perrs := recvPress(s.pool.Get(), game, player, stream)

	for {
		select {

		// process incoming messages from RedisPubSub, and send messages.
		case msg := <-msgs:
			if msg == nil {
				logger.Error(ctx, lc, "Message Channel has closed. Exiting.")
				return nil
			}

			logger.Info(ctx, lc, "Handling incoming messsage...")

			err := handle(con, game, player, stream, msg)
			if err != nil {
				// if we are EOF, then simply exit.
				if err == io.EOF {
					logger.Info(ctx, lc, "[Game] EOF. Closing connection.")
					return nil
				}
				return err
			}

		// check to see if there are any issues with press errors.
		case err := <-perrs:
			// remember, a closed channel, will return a nil err.
			if err != nil {
				logger.Error(ctx, lc, "There was a press error. %v", err)
				return err
			}
		}
	}
}

// pingRedis pings redis, to check if we are
// connected. Returns an error if there was a problem.
func (s *SimonSays) pingRedis() error {

	return backoff.Retry(func() error {
		con := s.pool.Get()
		defer con.Close()

		_, err := con.Do("PING")
		if err != nil {
			log.Printf("[Warn][Redis] Could not connect to Redis. %v", err)
		} else {
			log.Printf("[Info][Redis] Connected.")
		}

		return err
	}, backoff.NewExponentialBackOff())
}

func newPool(address string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", address)
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

// connectGame joins a game if one is in progress,
// or advertises this one as open if it is not.
func connectGame(ctx context.Context, con redis.Conn, game *Game, player *Request_Player, isNew bool) error {
	// if it's new, then add it for discovery.
	if isNew {
		err := addOpenGame(ctx, con, game)
		if err != nil {
			return err
		}
	} else {
		// shouldn't be any, so can be empty.
		b, err := game.EncodePresses()

		if err != nil {
			return err
		}

		// make sure we have 2 people subscribed at this point.
		if err := ensureSubscribers(ctx, con, game, 2); err != nil {
			return err
		}

		msg := message{Player: player.Id, Type: beginMessage, Data: b}

		err = publish(ctx, con, game, msg)
		if err != nil {
			return err
		}
	}

	return nil
}

// sendResponse Sends a request.
func sendResponse(stream SimonSays_GameServer, r *Response) error {
	lc := "Response"
	ctx := stream.Context()
	logger.Info(ctx, lc, "Sending response: %v", r)
	err := stream.Send(r)

	if err != nil {
		logger.Error(ctx, lc, "Error sending: %v", err)
	}

	return err
}

// receiveRequest receives a request.
func receiveRequest(stream SimonSays_GameServer) (*Request, error) {
	lc := "Request"
	ctx := stream.Context()
	req, err := stream.Recv()

	if err != nil {
		logger.Error(ctx, lc, "Error recieving: %v", err)
		return nil, err
	}

	logger.Info(ctx, lc, "Received: %v", req)

	return req, nil
}
