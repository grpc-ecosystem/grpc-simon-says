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
	"fmt"
	"io"

	"github.com/garyburd/redigo/redis"
	"github.com/grpc-simonsays/simonsays-server/simonsays/logger"
)

const (
	// begin the game
	beginMessage = "BEGIN"
	// stop a turn
	stopTurnMessage = "STOP_TURN"
	// light up a colour
	lightUpMessage = "LIGHTUP"
	// This player has lost
	lostMessage = "LOST_MESSAGE"
)

// Handler handles a Message that comes through redis pub/sub.
type handler func(redis.Conn, *Game, *Request_Player, SimonSays_GameServer, *message) error

// map of handlers for each message type
var handlers = map[string]handler{
	beginMessage:    beginHandler,
	stopTurnMessage: stopTurnHandler,
	lightUpMessage:  lightUpHandler,
	lostMessage:     lostHandler,
}

// handle Processing pub/sub events and does things with them
// Think "controller".
func handle(con redis.Conn, game *Game, player *Request_Player, stream SimonSays_GameServer, msg *message) error {
	lc := "Handler"
	ctx := stream.Context()
	logger.Info(ctx, lc, "Handling Message: %#v", msg)
	fn, ok := handlers[msg.Type]

	if !ok {
		logger.Error(ctx, lc, "Could not find a handler for this event. %#v", msg)
		return handlerNotFoundError("msg.Type")
	}

	return fn(con, game, player, stream, msg)
}

// beginHandler Streams BEGIN to client once we are good to go.
func beginHandler(con redis.Conn, game *Game, player *Request_Player, stream SimonSays_GameServer, msg *message) error {
	lc := "beginHandler"
	ctx := stream.Context()
	res := &Response{Event: &Response_Turn{Turn: Response_BEGIN}}

	err := sendResponse(stream, res)

	if err != nil {
		logger.Error(ctx, lc, "Error sending BEGIN event. %v", err)
		return err
	}

	// if not the player that BEGAN (so first player to join), then START your turn
	if msg.Player == player.Id {
		logger.Info(ctx, lc, "Publishing end turn %v", stopTurnMessage)
		return publish(ctx, con, game, message{Player: player.Id, Type: stopTurnMessage, Data: msg.Data})
	}

	logger.Info(ctx, lc, "Not doing anything with Begin. It's not my job.")
	return nil
}

// stopTurnHandler My turn has finished, so, tell the other player
// to START_TURN, and me to END_TURN.
func stopTurnHandler(con redis.Conn, game *Game, player *Request_Player, stream SimonSays_GameServer, msg *message) error {
	lc := "stopTurnHandler"
	ctx := stream.Context()

	// if I'm the player that sent out the message, let the client know
	if player.Id == msg.Player {
		return sendResponse(stream, &Response{Event: &Response_Turn{Turn: Response_STOP_TURN}})
	}

	// otherwise, it's time for the other player to start

	buf := bytes.NewBuffer(msg.Data)
	c := []Color{}
	err := gob.NewDecoder(buf).Decode(&c)

	if err != nil {
		logger.Error(ctx, lc, "Error decoding message colors. %#v. %v", msg, err)
		return err
	}

	logger.Info(ctx, lc, "Starting turn with colors: %v", c)
	game.StartTurn(c)
	return sendResponse(stream, &Response{Event: &Response_Turn{Turn: Response_START_TURN}})
}

// lightUpHandler handles LIGHTUP events, letting everyone know to lightup
// their colours.
func lightUpHandler(con redis.Conn, game *Game, player *Request_Player, stream SimonSays_GameServer, msg *message) error {
	lc := "lightUpHandler"
	ctx := stream.Context()
	c := new(Color)
	buf := bytes.NewBuffer(msg.Data)

	err := gob.NewDecoder(buf).Decode(c)

	if err != nil {
		logger.Error(ctx, lc, "Could not convert colour. %#v. %v", msg, err)
		return err
	}

	logger.Info(ctx, lc, "Sending stream response to Light Up %v", c)

	return sendResponse(stream, &Response{Event: &Response_Lightup{Lightup: *c}})
}

// what happens when the game is lost. Returns io.EOF to show that the game should be shut down.
func lostHandler(con redis.Conn, game *Game, player *Request_Player, stream SimonSays_GameServer, msg *message) error {
	lc := "lostHandler"
	ctx := stream.Context()

	logger.Info(ctx, lc, "Received Lost Event: %#v", msg)

	var err error

	// if I lost...
	turn := Response_WIN
	if msg.Player == player.Id {
		turn = Response_LOSE
	}

	err = sendResponse(stream, &Response{Event: &Response_Turn{Turn: turn}})
	if err != nil {
		return err
	}

	return io.EOF
}

type handlerNotFoundError string

// Error returns the string representation of a handlerNotFoundError.
func (h handlerNotFoundError) Error() string {
	return fmt.Sprintf("Could not find handler for Data event: %s", h)
}
