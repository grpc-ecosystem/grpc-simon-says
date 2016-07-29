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
	"testing"

	"github.com/garyburd/redigo/redis"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/context"
)

// TestAddOpenGame Tests adding an open game.
func TestAddOpenGame(t *testing.T) {
	Convey("When we have no open games", t, func() {
		server := mustSimonSays()
		defer server.Close()

		con := server.pool.Get()
		defer server.Close()
		_, err := con.Do("FLUSHDB")
		So(err, ShouldBeNil)

		result, err := redis.Values(con.Do("GET", openGames))

		So(err, ShouldEqual, redis.ErrNil)
		So(result, ShouldBeEmpty)

		ctx := context.TODO()

		Convey("We can add an open game", func() {
			game := NewGame("new game!!!")
			err := addOpenGame(ctx, con, game)
			So(err, ShouldBeNil)

			result, err := redis.Strings(con.Do("LRANGE", openGames, 0, -1))

			So(err, ShouldBeNil)
			So(result, ShouldNotBeEmpty)
			So(len(result), ShouldEqual, 1)
			So(result[0], ShouldEqual, game.ID)

			Convey("We can remove an open game", func() {
				err := closeOpenGame(ctx, con, game)
				So(err, ShouldBeNil)
				result, err := redis.Strings(con.Do("LRANGE", openGames, 0, -1))
				So(err, ShouldBeNil)
				So(result, ShouldBeEmpty)
			})
		})
	})
}

// TestFindGame Testing out if we can find a game.
func TestFindGame(t *testing.T) {

	Convey("When we have a simon says", t, func() {
		server := mustSimonSays()
		defer server.Close()
		con := server.pool.Get()

		_, err := con.Do("FLUSHDB")
		So(err, ShouldBeNil)

		ctx := context.TODO()

		Convey("And there is no game in the open games list, we should get a new game id", func() {
			gameid, isNewGame, err := findGame(context.TODO(), con)

			So(err, ShouldBeNil)
			So(gameid, ShouldNotBeNil)
			So(gameid, ShouldNotBeEmpty)
			So(isNewGame, ShouldBeTrue)
		})

		Convey("When there is an open game", func() {
			con := server.pool.Get()
			defer con.Close()

			game := NewGame("new game")
			err := addOpenGame(ctx, con, game)
			So(err, ShouldBeNil)

			foundGame, isNewGame, err := findGame(context.TODO(), con)
			So(err, ShouldBeNil)
			So(isNewGame, ShouldBeFalse)
			So(foundGame, ShouldResemble, game)
		})
	})

}
