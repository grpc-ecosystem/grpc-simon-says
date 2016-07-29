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
	"io"
	"testing"
	"time"

	uuid "github.com/nu7hatch/gouuid"
	. "github.com/smartystreets/goconvey/convey"
)

// TestRecvPress tests recieving a press.
func TestRecvPress(t *testing.T) {

	Convey("When you have a GREEN button being pressed", t, func() {
		stream := newMockStream()
		server := mustSimonSays()
		defer server.Close()

		player := &Request_Player{Id: "Player One"}

		con := server.pool.Get()

		_, err := con.Do("FLUSHDB")
		So(err, ShouldBeNil)
		u, err := uuid.NewV4()
		So(err, ShouldBeNil)
		game := NewGame(u.String())
		game.StartTurn([]Color{Color_GREEN})

		msgs, err := subscribe(stream.Context(), server.pool.Get(), game)
		So(err, ShouldBeNil)

		press := &Request{Event: &Request_Press{Press: Color_GREEN}}
		err = stream.PushRecv(press)
		So(err, ShouldBeNil)

		Convey("We should recieve a lightup event through pubsub", func() {
			errors := recvPress(server.pool.Get(), game, player, stream)

			select {
			case msg := <-msgs:
				So(msg.Type, ShouldEqual, lightUpMessage)

				// decode the data, make sure it's okay
				c := new(Color)
				buf := bytes.NewBuffer(msg.Data)
				err := gob.NewDecoder(buf).Decode(c)
				So(err, ShouldBeNil)
				So(*c, ShouldEqual, Color_GREEN)
				So(game.currentPresses, ShouldResemble, []Color{Color_GREEN})

			case <-time.After(3 * time.Second):
				So("Should recieve lightup event", ShouldBeNil)
			}

			Convey("We press one more color, we should switch turns", func() {
				press := &Request{Event: &Request_Press{Press: Color_GREEN}}
				err = stream.PushRecv(press)
				So(err, ShouldBeNil)
				stream.Close()

				// lightup msg
				select {
				case msg := <-msgs:
					So(msg.Type, ShouldEqual, lightUpMessage)
				case <-time.After(3 * time.Second):
					So("Should recieve lightup event", ShouldBeNil)
				}

				select {
				case msg := <-msgs:
					So(msg.Type, ShouldEqual, stopTurnMessage)

					colors := []Color{}
					// decode the data, make sure it's okay
					buf := bytes.NewBuffer(msg.Data)
					err := gob.NewDecoder(buf).Decode(&colors)
					So(err, ShouldBeNil)
					So(colors, ShouldResemble, []Color{Color_GREEN, Color_GREEN})

				case <-time.After(3 * time.Second):
					So("Should recieve StopTurnMessage", ShouldBeNil)
				}

			})

			// check for any errors.
			for err := range errors {
				if err != io.EOF {
					So(err, ShouldBeNil)
				}
			}
		})

	})
}

// testSendLightupEvent test out the send lightup event.
func TestSendLightupEvent(t *testing.T) {
	Convey("When you send a lightup event", t, func() {
		press := &Request_Press{Press: Color_GREEN}
		stream := newMockStream()
		server := mustSimonSays()
		defer server.Close()

		con := server.pool.Get()
		_, err := con.Do("FLUSHDB")
		So(err, ShouldBeNil)

		u, err := uuid.NewV4()
		So(err, ShouldBeNil)
		game := NewGame(u.String())

		msgs, err := subscribe(stream.Context(), server.pool.Get(), game)
		So(err, ShouldBeNil)

		err = sendLightupEvent(press, stream, con, game)
		So(err, ShouldBeNil)

		Convey("You should recieve a LightUpMessage over pubsub", func() {
			select {
			case msg := <-msgs:
				So(msg.Type, ShouldEqual, lightUpMessage)

				// decode the data, make sure it's okay
				c := new(Color)
				buf := bytes.NewBuffer(msg.Data)
				err := gob.NewDecoder(buf).Decode(c)
				So(err, ShouldBeNil)
				So(*c, ShouldEqual, Color_GREEN)

			case <-time.After(3 * time.Second):
				So("Should recieve lightup event", ShouldBeNil)
			}
		})
	})
}

// TestHandleEndOfTurn test out the handling of an end of turn.
func TestHandleEndOfTurn(t *testing.T) {
	Convey("When we have started a turn", t, func() {
		stream := newMockStream()
		server := mustSimonSays()
		defer server.Close()
		u, err := uuid.NewV4()
		So(err, ShouldBeNil)

		game := NewGame(u.String())
		player := &Request_Player{Id: "Player One"}

		con := server.pool.Get()
		colors := []Color{Color_GREEN, Color_BLUE}

		game.StartTurn(colors)

		msgs, err := subscribe(stream.Context(), server.pool.Get(), game)
		So(err, ShouldBeNil)

		Convey("We press a color that is right", func() {
			err := game.PressColor(Color_GREEN)
			So(err, ShouldBeNil)

			lost, err := handleEndOfTurn(stream, con, game, player)
			So(err, ShouldBeNil)
			So(lost, ShouldBeFalse)

			select {
			case msg := <-msgs:
				So(msg, ShouldBeNil)
			default:
				// do nothing - there should be no message.
			}

			Convey("Press another color that's correct", func() {
				err := game.PressColor(Color_BLUE)
				So(err, ShouldBeNil)

				lost, err := handleEndOfTurn(stream, con, game, player)
				So(err, ShouldBeNil)
				So(lost, ShouldBeFalse)

				select {
				case msg := <-msgs:
					So(msg, ShouldBeNil)
				default:
					// do nothing - there should be no message.
				}

				Convey("Press the colour that in the next color input", func() {
					err := game.PressColor(Color_BLUE)
					So(err, ShouldBeNil)

					So(game.IsMyTurn(), ShouldBeFalse)

					lost, err := handleEndOfTurn(stream, con, game, player)
					So(err, ShouldBeNil)
					So(lost, ShouldBeFalse)

					select {
					case msg := <-msgs:
						So(msg.Type, ShouldEqual, stopTurnMessage)

						colors := []Color{}
						// decode the data, make sure it's okay.
						buf := bytes.NewBuffer(msg.Data)
						err := gob.NewDecoder(buf).Decode(&colors)
						So(err, ShouldBeNil)
						So(colors, ShouldResemble, []Color{Color_GREEN, Color_BLUE, Color_BLUE})

					case <-time.After(3 * time.Second):
						So("Should recieve StopTurnMessage", ShouldBeNil)
					}
				})
			})

			Convey("Press a color input that is incorrect", func() {
				err := game.PressColor(Color_YELLOW)
				So(err, ShouldBeNil)

				lost, err := handleEndOfTurn(stream, con, game, player)
				So(err, ShouldBeNil)
				So(lost, ShouldBeTrue)
				So(game.IsMyTurn(), ShouldBeFalse)

				select {
				case msg := <-msgs:
					So(msg.Type, ShouldEqual, lostMessage)
					So(msg.Player, ShouldEqual, player.Id)
				case <-time.After(3 * time.Second):
					So("Should recieve LostMessage", ShouldBeNil)
				}
			})
		})
	})
}
