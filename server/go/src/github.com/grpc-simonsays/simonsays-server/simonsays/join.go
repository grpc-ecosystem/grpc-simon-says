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
	"github.com/garyburd/redigo/redis"
	"github.com/grpc-simonsays/simonsays-server/simonsays/logger"
	uuid "github.com/nu7hatch/gouuid"
	"golang.org/x/net/context"
)

const openGames = "OpenGames"

// findGame finds a game in the list of open games. If one doesn't exist, creates a new gameid
// returns a new Game and if it's a new game or not.
func findGame(ctx context.Context, con redis.Conn) (*Game, bool, error) {
	lc := "FindGame"

	// do we have an open game?
	gameID, err := redis.String(con.Do("RPOP", openGames))

	// ignore nil errors, since that is expected
	if err != nil && err != redis.ErrNil {
		logger.Error(ctx, lc, "Error finding open game: %v", err)
		return new(Game), false, err
	}

	// is this a brand new game?
	isNew := (gameID == "")

	if isNew {
		logger.Info(ctx, lc, "Could not find open game, creating one... ")
		u, err := uuid.NewV4()
		if err != nil {
			return nil, false, err
		}
		gameID = u.String()
	}

	return NewGame(gameID), isNew, nil
}

// addOpenGame Adds an open game to the list.
func addOpenGame(ctx context.Context, con redis.Conn, g *Game) error {
	logger.Info(ctx, "AddOpenGame", "Adding open game %v", g.ID)
	_, err := con.Do("LPUSH", openGames, g.ID)
	return err
}

// closeOpenGame make sure the open game is removed
// from the open game list.
func closeOpenGame(ctx context.Context, con redis.Conn, g *Game) error {
	logger.Info(ctx, "CloseOpenGame", "Removing open game %v", g.ID)
	_, err := con.Do("LREM", openGames, 1, g.ID)
	return err
}
