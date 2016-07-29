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
	"fmt"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/grpc-simonsays/simonsays-server/simonsays/logger"
	"golang.org/x/net/context"
)

// message is the pubsub message
// that gets sent.
type message struct {
	//leave these as exported, so gob will encode/decode them
	Type   string
	Player string
	Data   []byte
}

// encode convert into []bytes as gob
func (m *message) marshalGob() ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(m)
	return buf.Bytes(), err
}

// decode converts from []bytes back into the object
func (m *message) unmarshalGob(b []byte) error {
	buf := bytes.NewBuffer(b)
	dec := gob.NewDecoder(buf)
	return dec.Decode(m)
}

// subscribe subscribes to the topic for this game
// Returns a channel of Messages that can be used to receive messages
// Make sure to send it a new redis.Conn. It will handle closing it
// when finished.
func subscribe(ctx context.Context, con redis.Conn, g *Game) (<-chan *message, error) {
	lc := "Subscribe"
	c := make(chan *message)

	psc := redis.PubSubConn{con}

	logger.Info(ctx, lc, "Subscribing to topic '%v'", g.ID)

	err := psc.Subscribe(g.ID)
	if err != nil {
		logger.Info(ctx, lc, "Error Subscribing. %v", err)
		close(c)
		return c, err
	}

	go func(c chan<- *message) {
		defer con.Close()
		defer close(c)
		for {
			switch v := psc.Receive().(type) {
			case redis.Message:
				msg := new(message)
				err := msg.unmarshalGob(v.Data)
				if err != nil {
					logger.Error(ctx, lc, "Could not decode message. Closing channel. %v, %v", v, err)
					return
				}

				logger.Info(ctx, lc, "Received Message. Sending %#v to channel.", msg)

				c <- msg

				// special case for LostMessage, since we know to close at this point.
				if msg.Type == lostMessage {
					logger.Info(ctx, lc, "Lost Message, Closing Subscribe Pipeline.")
					return
				}
			case error:
				logger.Error(ctx, lc, "Error processing messages. Closing channel. %v", err)
				return
			default:
				logger.Info(ctx, lc, "Received unknown message. Ignored: %#v", v)
			}
		}
	}(c)

	return c, nil
}

// publish publishes a message to the game's topic.
func publish(ctx context.Context, con redis.Conn, g *Game, msg message) error {
	lc := "Publish"

	logger.Info(ctx, lc, "Sending message: %#v, to topic: '%v'", msg, g.ID)

	data, err := msg.marshalGob()

	if err != nil {
		logger.Error(ctx, lc, "Error encoding message. %#v, %v", msg, err)
		return err
	}

	_, err = con.Do("PUBLISH", g.ID, data)

	if err != nil {
		logger.Error(ctx, lc, "Error publishing message. %#v, %v", msg, err)
	}

	return err
}

// ensureSubscribers Make sure n number of Game subscriptions at this point.
// Blocks until we have two people. Times out on too many retries.
func ensureSubscribers(ctx context.Context, con redis.Conn, g *Game, n int) error {
	lc := "EnsureSubscribers"

	for i := 0; i <= 5; i++ {
		res, err := con.Do("PUBSUB", "NUMSUB", g.ID)

		if err != nil {
			logger.Error(ctx, lc, "Error getting number of subscriptions: %v", err)
			return err
		}

		vals, err := redis.Values(res, err)

		if err != nil {
			logger.Error(ctx, lc, "Error converting to values: %v", err)
			return err
		}

		if l := len(vals); l != 2 {
			err := fmt.Errorf("Should only be two items in the result. Weird. %#v. %v.", vals, l)
			logger.Error(ctx, lc, err.Error())
			return err
		}

		count, ok := vals[1].(int64)

		if !ok {
			val := vals[1]
			err := fmt.Errorf("Second value should be an integer. %T %#v", val, val)
			logger.Error(ctx, lc, err.Error())
			return err
		}

		logger.Info(ctx, lc, "Found %v subscriptions for Game. Require %v", count, n)

		if int(count) == n {
			return nil
		}

		logger.Info(ctx, lc, "Could not find enough subscriptions, retrying...")
		time.Sleep(100 * time.Millisecond)
	}

	err := errors.New("Timeout attempting to ensure subscriber count of " + strconv.Itoa(n))
	logger.Error(ctx, lc, err.Error())
	return err
}
