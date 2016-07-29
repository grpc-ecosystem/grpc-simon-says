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
	"bytes"
	"encoding/gob"
	"errors"

	"github.com/garyburd/redigo/redis"
	"github.com/grpc-simonsays/simonsays-server/simonsays/logger"
)

// recvPress Manages receiving Press Events through a go-routine.
// Will need it's own Redis connection, that it will close once complete.
// Sends io.EOF when the connection closes, and pushes the error into the chan if it
// occurs.
func recvPress(con redis.Conn, game *Game, player *Request_Player, stream SimonSays_GameServer) <-chan error {
	lc := "RecvPress"
	ctx := stream.Context()
	c := make(chan error, 10)

	logger.Info(ctx, lc, "Start recieving Press events...")

	go func() {
		defer close(c)
		defer con.Close()
		for {
			stop, err := handleColorPress(con, game, player, stream)
			if err != nil {
				c <- err
				return
			} else if stop {
				return
			}
		}
	}()

	return c
}

// handleColorPress handles one color being pressed.
// If it's the player turn it modifies the given game and sends a lightUpMessage to Redis.
// This function is thread safe.
func handleColorPress(con redis.Conn, game *Game, player *Request_Player, stream SimonSays_GameServer) (bool, error) {
	lc := "handleColorPress"
	ctx := stream.Context()
	press, err := receivePressRequest(stream)

	if err != nil {
		logger.Error(ctx, lc, "Press Error. Sending to err channel, and shutting down %v", err)
		return true, err
	}

	logger.Info(ctx, lc, "Press Received: %v", press)

	//lock the game for this entire block, since we are doing lots of things with
	//it, and this will prevent any concurrency issues.
	game.mu.Lock()
	defer game.mu.Unlock()

	// only accept input when it is my turn!
	if !game.isMyTurn() {
		logger.Info(ctx, lc, "Not my turn. Ignored press.")
		return false, nil
	}

	err = game.pressColor(press.Press)
	if err == ErrColorPressedOutOfTurn {
		logger.Info(ctx, lc, "Colour pressed out of turn. Ignored.")
	} else if err != nil {
		return true, err
	}

	err = sendLightupEvent(press, stream, con, game)
	if err != nil {
		return true, err
	}

	// When you reach the point that the game has turned.
	return handleEndOfTurn(stream, con, game, player)
}

// receivePress Receives a press. Returns an error if there is an issue, and publishes it
// to redis as well.
func receivePressRequest(stream SimonSays_GameServer) (*Request_Press, error) {
	lc := "receivePressRequest"
	ctx := stream.Context()
	res, err := receiveRequest(stream)

	if err != nil {
		logger.Error(ctx, lc, "Error recieving from gRPC: %v", err)
		return nil, err
	}

	press, ok := res.Event.(*Request_Press)

	if !ok {
		err := errors.New("Recieved a request other than a Press")
		logger.Error(ctx, lc, "Error: %v. %v", err, res)
		return nil, err
	}

	return press, nil
}

// sendLightupEvent sends out a lightup event to everyone.
func sendLightupEvent(press *Request_Press, stream SimonSays_GameServer, con redis.Conn, game *Game) error {
	buf := new(bytes.Buffer)
	err := gob.NewEncoder(buf).Encode(press.Press)
	ctx := stream.Context()

	if err != nil {
		logger.Info(ctx, "sendLightupEvent", "Error gob encoding Color. %v, %v", press.Press, err)
		return err
	}

	// send out lightup events.
	return publish(ctx, con, game, message{Type: lightUpMessage, Data: buf.Bytes()})
}

// handleEndOfTurn handles if it is the end of the turn, and if the player has lost (bool).
func handleEndOfTurn(stream SimonSays_GameServer, con redis.Conn, game *Game, player *Request_Player) (bool, error) {
	lc := "handleEndOfTurn"
	ctx := stream.Context()

	// if not my turn, exit early.
	if game.isMyTurn() {
		return false, nil
	}

	if game.match() {
		b, err := game.encodePresses()
		if err != nil {
			logger.Error(ctx, lc, "Error encoding presses: %#v, %v", game, err)
			return false, err
		}

		msg := message{Type: stopTurnMessage, Player: player.Id, Data: b}
		if err := publish(ctx, con, game, msg); err != nil {
			logger.Error(ctx, lc, "error publishing StopTurnMessage %#v, %v", msg, err)
			return false, err
		}

		return false, nil
	}

	// if there is no match, you did something wrong. otherwise, my friend, you have lost the game.
	msg := message{Type: lostMessage, Player: player.Id}
	if err := publish(ctx, con, game, msg); err != nil {
		logger.Error(ctx, lc, "error publishing LostMessage %#v, %v", msg, err)
		return false, err
	}

	// and we are done taking input - end of game!
	logger.Info(ctx, lc, "We are done taking input. Returning that we have lost. %#v", game)
	return true, nil
}
