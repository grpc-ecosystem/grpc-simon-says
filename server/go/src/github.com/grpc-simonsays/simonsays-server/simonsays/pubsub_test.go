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
	"time"

	uuid "github.com/nu7hatch/gouuid"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/context"
)

// TestEncodeDecode testing gob encode/decode.
func TestEncodeDecode(t *testing.T) {
	Convey("When you have a message", t, func() {
		msg := &message{Player: "Player One", Data: []byte("My Data"), Type: "BEGIN"}

		Convey("You can encode it", func() {
			b, err := msg.marshalGob()
			So(err, ShouldBeNil)
			So(b, ShouldNotBeEmpty)
			Convey("And then decode it", func() {
				msg2 := new(message)
				err := msg2.unmarshalGob(b)
				So(err, ShouldBeNil)
				So(msg2, ShouldResemble, msg)
			})
		})
	})
}

// TestPublishAndSubscribe Testing out the Message Publish and Subscribe mechanism.
func TestPublishAndSubscribe(t *testing.T) {
	Convey("When you are subscribed to a topic", t, func() {
		server := mustSimonSays()
		defer server.Close()

		con := server.pool.Get()
		_, err := con.Do("FLUSHDB")
		So(err, ShouldBeNil)

		u, err := uuid.NewV4()
		So(err, ShouldBeNil)
		game := NewGame(u.String())

		ctx := context.TODO()

		pubsub, err := subscribe(ctx, server.pool.Get(), game)
		So(err, ShouldBeNil)
		Convey("We can create a message", func() {
			msg := message{Player: "Player One", Data: []byte("BEGIN!"), Type: "BEGIN"}

			Convey("We can publish a message to the topic", func() {
				err := publish(ctx, con, game, msg)
				So(err, ShouldBeNil)

				Convey("And we can retrieve it back", func() {

					select {
					case msg2 := <-pubsub:
						So(msg2, ShouldNotBeNil)
						So(msg2, ShouldResemble, &msg)
					case <-time.After(5 * time.Second):
						So("Timeout getting message", ShouldBeNil)
					}

				})
			})
		})
	})
}

// TestEnsureSubscribers test out ensuring we have the right number of subscribers.
func TestEnsureSubscribers(t *testing.T) {

	Convey("When you have a game that you can subscribe to", t, func(c C) {
		server := mustSimonSays()
		defer server.Close()

		u, err := uuid.NewV4()
		So(err, ShouldBeNil)
		game := NewGame(u.String())
		ctx := context.TODO()

		done := make(chan bool)

		go func(c C) {
			defer close(done)
			err := ensureSubscribers(ctx, server.pool.Get(), game, 2)
			c.So(err, ShouldBeNil)
		}(c)

		select {
		case <-done:
			So("Should not be done at this point. No subscribers", ShouldBeNil)
		case <-time.After(100 * time.Millisecond):
			// this is what should happen.
		}

		_, err = subscribe(ctx, server.pool.Get(), game)
		So(err, ShouldBeNil)

		select {
		case <-done:
			So("Should not be done at this point. Only 1 subscriber", ShouldBeNil)
		case <-time.After(100 * time.Millisecond):
			// this is what should happen.
		}

		_, err = subscribe(ctx, server.pool.Get(), game)
		So(err, ShouldBeNil)

		select {
		case <-done:
			// this should work now
		case <-time.After(500 * time.Millisecond):
			So("We have two subscribers now, so done should close", ShouldBeNil)
		}
	})

}
