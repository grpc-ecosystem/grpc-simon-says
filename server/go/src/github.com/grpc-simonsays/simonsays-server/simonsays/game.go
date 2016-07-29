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
	"sync"
)

// Game represents a Game that an individual player is playing.
// It keeps track of internal Game state for that player.
// Only use exported methods are guaranteed to be concurrently safe.
type Game struct {
	ID             string
	currentPresses []Color
	validPresses   []Color
	myTurn         bool
	mu             sync.RWMutex
}

// ErrColorPressedOutOfTurn is returned when a colour is pressed outside
// of the Game player's turn.
var ErrColorPressedOutOfTurn = errors.New("Color pressed outside of player turn")

// NewGame returns a new Game for a player
func NewGame(id string) *Game {
	return &Game{
		ID:     id,
		myTurn: false,
	}
}

// StartTurn starts the player's turn. It is passed the sequence of Colors the player
// needs to match during this turn to continue to the next round.
func (g *Game) StartTurn(p []Color) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.validPresses = p
	g.currentPresses = nil
	g.myTurn = true
}

// pressColor is an unlocked version of PressColor.
func (g *Game) pressColor(c Color) error {
	if !g.myTurn {
		return ErrColorPressedOutOfTurn
	}

	g.currentPresses = append(g.currentPresses, c)

	if len(g.currentPresses)-1 == len(g.validPresses) || !g.match() {
		g.myTurn = false
	}

	return nil
}

// PressColor should be called when this player presses a colour.
// This will append the colour value to the list of currentPresses.
// Returns a ErrColorPressedOutOfTurn if it not this player's turn
// Will set IsMyTurn() to false when the number of currentPresses is
// one more than the current validPresses (what colours the last player pressed) value,
// or colours don't match up to the previous turn's colour sequence.
func (g *Game) PressColor(c Color) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.pressColor(c)
}

// encodePresses is the unlocked version of EncodePresses.
func (g *Game) encodePresses() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := gob.NewEncoder(buf).Encode(g.currentPresses)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// EncodePresses encodes the current set of Color
// presses into an []bytes for transmission.
func (g *Game) EncodePresses() ([]byte, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.encodePresses()
}

// isMyTurn is a unlocked version if IsMyTurn.
func (g *Game) isMyTurn() bool {
	return g.myTurn
}

// IsMyTurn Is this this player's turn?
func (g *Game) IsMyTurn() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.isMyTurn()
}

// Match checks to see if the currently entered presses, line up with
// what we have as valid presses.
func (g *Game) Match() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.match()
}

// match is an unlocked version of Match().
func (g *Game) match() bool {

	//only check presses, up to whichever list is shortest
	l := len(g.validPresses)
	if n := len(g.currentPresses); n < l {
		l = n
	}

	for i, v := range g.currentPresses[:l] {
		if v != g.validPresses[i] {
			return false
		}
	}

	return true
}
