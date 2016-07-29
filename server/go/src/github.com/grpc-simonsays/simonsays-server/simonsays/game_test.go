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

	. "github.com/smartystreets/goconvey/convey"
)

// TestGame tests out the game state for a full game.
func TestGame(t *testing.T) {
	Convey("When you have a game", t, func() {
		gameID := "My New Game"
		game := NewGame(gameID)
		So(game.ID, ShouldEqual, gameID)
		So(game.Match(), ShouldBeTrue)
		So(game.IsMyTurn(), ShouldBeFalse)

		Convey("and we have an initial set of colours", func() {
			colors := []Color{Color_GREEN, Color_BLUE}

			Convey("We can start a turn", func() {
				game.StartTurn(colors)
				So(game.IsMyTurn(), ShouldBeTrue)
				So(game.Match(), ShouldBeTrue)

				Convey("We can press the first right colour", func() {
					err := game.PressColor(colors[0])
					So(err, ShouldBeNil)
					So(game.IsMyTurn(), ShouldBeTrue)
					So(game.Match(), ShouldBeTrue)

					Convey("and the second right color", func() {
						err := game.PressColor(colors[1])
						So(err, ShouldBeNil)
						So(game.IsMyTurn(), ShouldBeTrue)
						So(game.Match(), ShouldBeTrue)

						Convey("and the third new color", func() {
							err := game.PressColor(Color_YELLOW)
							So(err, ShouldBeNil)
							So(game.IsMyTurn(), ShouldBeFalse)
							So(game.Match(), ShouldBeTrue)

							Convey("and a fourth extra Color should be ignored", func() {
								err := game.PressColor(Color_YELLOW)
								So(err, ShouldEqual, ErrColorPressedOutOfTurn)
								So(game.IsMyTurn(), ShouldBeFalse)
								So(game.Match(), ShouldBeTrue)

								Convey("We can start a second new turn, and the state should reset", func() {
									colors := []Color{Color_GREEN, Color_RED}
									game.StartTurn(colors)
									So(game.IsMyTurn(), ShouldBeTrue)
									So(game.Match(), ShouldBeTrue)

									Convey("We can press the first right colour", func() {
										err := game.PressColor(colors[0])
										So(err, ShouldBeNil)
										So(game.IsMyTurn(), ShouldBeTrue)
										So(game.Match(), ShouldBeTrue)

										Convey("and the second right color", func() {
											err := game.PressColor(colors[1])
											So(err, ShouldBeNil)
											So(game.IsMyTurn(), ShouldBeTrue)
											So(game.Match(), ShouldBeTrue)

											Convey("and the third new color", func() {
												err := game.PressColor(Color_YELLOW)
												So(err, ShouldBeNil)
												So(game.IsMyTurn(), ShouldBeFalse)
												So(game.Match(), ShouldBeTrue)

												Convey("and a fourth extra Color should be ignored", func() {
													err := game.PressColor(Color_YELLOW)
													So(err, ShouldEqual, ErrColorPressedOutOfTurn)
													So(game.IsMyTurn(), ShouldBeFalse)
													So(game.Match(), ShouldBeTrue)

												})
											})
										})
									})

								})

							})
						})
					})
				})

				Convey("We can press a wrong first color", func() {
					err := game.PressColor(Color_YELLOW)
					So(err, ShouldBeNil)
					So(game.IsMyTurn(), ShouldBeFalse)
					So(game.Match(), ShouldBeFalse)

					Convey("and the second right color, should be ignored", func() {
						err := game.PressColor(colors[1])
						So(err, ShouldEqual, ErrColorPressedOutOfTurn)
						So(game.IsMyTurn(), ShouldBeFalse)
						So(game.Match(), ShouldBeFalse)
					})
				})

			})
		})

	})
}

// TestEncodePresses make sure we can encode this for passing around.
func TestEncodePresses(t *testing.T) {
	Convey("When you have a game", t, func() {
		game := NewGame("hello world")
		colors := []Color{Color_GREEN, Color_BLUE}
		game.StartTurn(colors)

		Convey("And you press some colours", func() {
			for _, c := range colors {
				err := game.PressColor(c)
				So(err, ShouldBeNil)
			}

			Convey("The encoded presses should match the actual colors", func() {
				b, err := game.EncodePresses()
				So(err, ShouldBeNil)
				buf := bytes.NewBuffer(b)

				result := new([]Color)
				err = gob.NewDecoder(buf).Decode(result)
				So(err, ShouldBeNil)

				So(result, ShouldResemble, &colors)
			})
		})
	})
}
