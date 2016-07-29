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
	"fmt"
	"sync"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

// TestRedisConn tests that we can connect to Redis.
func TestRedisConn(t *testing.T) {
	Convey("When we have a Simon Says", t, func() {
		game, err := NewSimonSays("")
		So(err, ShouldBeNil)
		So(game, ShouldNotBeNil)
		Convey("We can ping redis successfully", func() {
			con := game.pool.Get()
			defer con.Close()

			res, err := con.Do("PING")
			So(err, ShouldBeNil)
			So("PONG", ShouldEqual, res)
		})
	})
}

// TestSimpleGame tests out a very basic game, with only
// one colour being pressed, before loseing.
func TestSimpleGame(t *testing.T) {

	playerOne := newMockStream()
	playerTwo := newMockStream()

	Convey("Given a SimonSays", t, func() {
		game, err := NewSimonSays("")
		So(err, ShouldBeNil)
		So(game, ShouldNotBeNil)
		defer game.Close()

		Convey("We should be able to complete a simple game", func(c C) {
			wg := sync.WaitGroup{}

			// player one joins
			err := playerOne.PushRecv(&Request{Event: &Request_Join{Join: &Request_Player{Id: "Player One"}}})
			So(err, ShouldBeNil)

			// join, and wait
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := game.Game(playerOne)
				c.So(err, ShouldBeNil)
			}()

			// player two joins
			err = playerTwo.PushRecv(&Request{Event: &Request_Join{Join: &Request_Player{Id: "Player Two"}}})
			So(err, ShouldBeNil)

			wg.Add(1)
			// hacky, but can't think of a better way. Wait Group helps, but only sometimes.
			time.Sleep(time.Second)
			go func() {
				defer wg.Done()
				err := game.Game(playerTwo)
				c.So(err, ShouldBeNil)
			}()

			// do the more complicate test here first, just to be sure.
			res, err := playerOne.PullSend()
			So(err, ShouldBeNil)
			turn, ok := res.Event.(*Response_Turn)
			So(ok, ShouldBeTrue)
			So(turn.Turn, ShouldEqual, Response_BEGIN)

			res, err = playerTwo.PullSend()
			So(err, ShouldBeNil)
			turn, ok = res.Event.(*Response_Turn)
			So(ok, ShouldBeTrue)
			So(turn.Turn, ShouldEqual, Response_BEGIN)

			// player one should receive a TURN_START
			So(playerOne, shouldState, Response_START_TURN)

			// player two should receive a TURN_STOP
			So(playerTwo, shouldState, Response_STOP_TURN)

			// play one should send a GREEN
			mustPress(playerOne, Color_GREEN)

			// all should receive a lightup GREEN
			So(playerOne, shouldLightup, Color_GREEN)
			So(playerTwo, shouldLightup, Color_GREEN)

			// player one should receive a TURN_STOP
			So(playerOne, shouldState, Response_STOP_TURN)
			// player two should receive a TURN_START
			So(playerTwo, shouldState, Response_START_TURN)

			// player two enters BLUE
			mustPress(playerTwo, Color_BLUE)

			// all players should light up blue
			So(playerOne, shouldLightup, Color_BLUE)
			So(playerTwo, shouldLightup, Color_BLUE)

			// player one should get a WIN
			So(playerOne, shouldState, Response_WIN)

			// player two should get a LOSE
			So(playerTwo, shouldState, Response_LOSE)

			wg.Wait()
		})
	})
}

// TestMoreComplexGame tests more complex game, with more than
// one colour being pressed before the game is over.
func TestMoreComplexGame(t *testing.T) {

	playerOne := newMockStream()
	playerTwo := newMockStream()

	Convey("Given a SimonSays", t, func() {
		game, err := NewSimonSays("")
		So(err, ShouldBeNil)
		So(game, ShouldNotBeNil)
		defer game.Close()

		Convey("We should be able to complete a slightly more complex game", func(c C) {
			wg := sync.WaitGroup{}

			// player one joins
			err := playerOne.PushRecv(&Request{Event: &Request_Join{Join: &Request_Player{Id: "Player One"}}})
			So(err, ShouldBeNil)

			// join, and wait
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := game.Game(playerOne)
				c.So(err, ShouldBeNil)
			}()

			// player two joins
			err = playerTwo.PushRecv(&Request{Event: &Request_Join{Join: &Request_Player{Id: "Player Two"}}})
			So(err, ShouldBeNil)

			wg.Add(1)
			// hacky, but can't think of a better way. Wait Group helps, but only sometimes.
			time.Sleep(time.Second)
			go func() {
				defer wg.Done()
				err := game.Game(playerTwo)
				c.So(err, ShouldBeNil)
			}()

			So(playerOne, shouldState, Response_BEGIN)
			So(playerTwo, shouldState, Response_BEGIN)

			So(playerOne, shouldState, Response_START_TURN)
			So(playerTwo, shouldState, Response_STOP_TURN)

			mustPress(playerOne, Color_GREEN)
			So(playerOne, shouldLightup, Color_GREEN)
			So(playerTwo, shouldLightup, Color_GREEN)

			So(playerOne, shouldState, Response_STOP_TURN)
			So(playerTwo, shouldState, Response_START_TURN)

			mustPress(playerTwo, Color_GREEN)
			So(playerOne, shouldLightup, Color_GREEN)
			So(playerTwo, shouldLightup, Color_GREEN)

			mustPress(playerTwo, Color_BLUE)
			So(playerOne, shouldLightup, Color_BLUE)
			So(playerTwo, shouldLightup, Color_BLUE)

			So(playerOne, shouldState, Response_START_TURN)
			So(playerTwo, shouldState, Response_STOP_TURN)

			mustPress(playerOne, Color_GREEN)
			So(playerOne, shouldLightup, Color_GREEN)
			So(playerTwo, shouldLightup, Color_GREEN)

			mustPress(playerOne, Color_BLUE)
			So(playerOne, shouldLightup, Color_BLUE)
			So(playerTwo, shouldLightup, Color_BLUE)

			mustPress(playerOne, Color_YELLOW)
			So(playerOne, shouldLightup, Color_YELLOW)
			So(playerTwo, shouldLightup, Color_YELLOW)

			So(playerOne, shouldState, Response_STOP_TURN)
			So(playerTwo, shouldState, Response_START_TURN)

			// now the playerTwo will fail

			mustPress(playerTwo, Color_GREEN)
			So(playerOne, shouldLightup, Color_GREEN)
			So(playerTwo, shouldLightup, Color_GREEN)

			// this one is wrong
			mustPress(playerTwo, Color_GREEN)
			So(playerOne, shouldLightup, Color_GREEN)
			So(playerTwo, shouldLightup, Color_GREEN)

			So(playerOne, shouldState, Response_WIN)
			So(playerTwo, shouldState, Response_LOSE)

			wg.Wait()
		})
	})
}

func shouldLightup(player interface{}, args ...interface{}) string {
	res, err := player.(*mockStream).PullSend()
	if err != nil {
		return fmt.Sprintf("Error should be nil. %v", err)
	}

	lu, ok := res.Event.(*Response_Lightup)

	if !ok {
		return fmt.Sprintf("Response not a lightup event: %v", res)
	}

	if lu.Lightup != args[0].(Color) {
		return fmt.Sprintf("Lightup %v != Color: %v", lu.Lightup, args[0])
	}

	return ""
}

func shouldState(player interface{}, args ...interface{}) string {
	res, err := player.(*mockStream).PullSend()
	if err != nil {
		return fmt.Sprintf("Error should be nil. %v", err)
	}

	t, ok := res.Event.(*Response_Turn)

	if !ok {
		return fmt.Sprintf("Response not a state event: %v", res)
	}

	if t.Turn != args[0].(Response_State) {
		return fmt.Sprintf("Turn State %v != Response Turn: %v", t.Turn, args[0])
	}

	return ""
}

// pressColor press a colour for testing
func mustPress(p *mockStream, c Color) {
	// a push should never really panic. This would be really bad and weird.
	err := p.PushRecv(&Request{Event: &Request_Press{Press: c}})
	if err != nil {
		panic(err)
	}
}
