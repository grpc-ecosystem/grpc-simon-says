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
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

// TestBeginHandler test the begin handler.
func TestBeginHandler(t *testing.T) {
	stream := newMockStream()
	server := mustSimonSays()
	defer server.Close()

	Convey("When you have a begin event", t, func() {
		player := &Request_Player{Id: "Player One"}
		data := []byte("This is my data")
		msg := &message{Type: beginMessage, Data: data}
		game := NewGame("game one")

		c, err := subscribe(stream.Context(), server.pool.Get(), game)
		So(err, ShouldBeNil)

		Convey("And it's not your player sending the event", func() {
			err = beginHandler(server.pool.Get(), game, player, stream, msg)
			So(err, ShouldBeNil)

			res, err := stream.PullSend()
			So(err, ShouldBeNil)

			turn, ok := res.Event.(*Response_Turn)
			So(ok, ShouldBeTrue)
			So(turn.Turn, ShouldEqual, Response_BEGIN)

			select {
			case <-c:
				So("Should be no message", ShouldBeNil)
			case <-time.After(2 * time.Second):
			}
		})

		Convey("and the player is sending the event", func() {
			msg := &message{Type: beginMessage, Player: player.Id, Data: data}
			err = beginHandler(server.pool.Get(), game, player, stream, msg)
			So(err, ShouldBeNil)

			res, err := stream.PullSend()
			So(err, ShouldBeNil)

			turn, ok := res.Event.(*Response_Turn)
			So(ok, ShouldBeTrue)
			So(turn.Turn, ShouldEqual, Response_BEGIN)

			select {
			case msg := <-c:
				So(msg.Player, ShouldEqual, "Player One")
				So(msg.Type, ShouldEqual, stopTurnMessage)
				So(msg.Data, ShouldResemble, data)
			case <-time.After(2 * time.Second):
				So(true, ShouldBeNil)
			}
		})

	})
}

// testStopTurnHandler test the begin handler.
func TestStopTurnHandler(t *testing.T) {
	stream := newMockStream()
	server := mustSimonSays()
	defer server.Close()

	Convey("When you have a stop turn event", t, func() {
		player := &Request_Player{Id: "Player One"}
		cols := []Color{Color_GREEN, Color_BLUE}
		buf := new(bytes.Buffer)
		err := gob.NewEncoder(buf).Encode(cols)
		So(err, ShouldBeNil)

		msg := &message{Type: stopTurnMessage, Data: buf.Bytes()}
		game := NewGame("game one")
		So(game.IsMyTurn(), ShouldBeFalse)

		Convey("And it's not your player sending the event", func() {
			err := stopTurnHandler(server.pool.Get(), game, player, stream, msg)
			So(err, ShouldBeNil)

			res, err := stream.PullSend()
			So(err, ShouldBeNil)

			turn, ok := res.Event.(*Response_Turn)
			So(ok, ShouldBeTrue)
			So(turn.Turn, ShouldEqual, Response_START_TURN)
			So(game.IsMyTurn(), ShouldBeTrue)
		})

		Convey("and the player is sending the event", func() {
			msg := &message{Type: stopTurnMessage, Player: player.Id, Data: buf.Bytes()}
			err := stopTurnHandler(server.pool.Get(), game, player, stream, msg)
			So(err, ShouldBeNil)

			res, err := stream.PullSend()
			So(err, ShouldBeNil)

			turn, ok := res.Event.(*Response_Turn)
			So(ok, ShouldBeTrue)
			So(turn.Turn, ShouldEqual, Response_STOP_TURN)
			So(game.IsMyTurn(), ShouldBeFalse)
		})

	})
}

// TestLightUpHandler test the lightup handler.
func TestLightUpHandler(t *testing.T) {
	stream := newMockStream()
	server := mustSimonSays()
	defer server.Close()

	Convey("When you have a Lightup event", t, func() {
		game := NewGame("game one")
		player := &Request_Player{Id: "Player One"}
		buf := new(bytes.Buffer)
		err := gob.NewEncoder(buf).Encode(Color_GREEN)
		So(err, ShouldBeNil)

		msg := &message{Type: lightUpMessage, Data: buf.Bytes()}

		Convey("we should recieve a Lightup gRPC message", func() {
			err := lightUpHandler(server.pool.Get(), game, player, stream, msg)
			So(err, ShouldBeNil)

			res, err := stream.PullSend()
			So(err, ShouldBeNil)
			So(res.GetLightup(), ShouldEqual, Color_GREEN)
		})
	})
}
