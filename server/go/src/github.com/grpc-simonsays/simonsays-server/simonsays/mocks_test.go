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
	"errors"
	"io"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

//timeOut value for retrieving data from the mockStream channels.
//timeouts should only happen when tests fail.
const timeOut = 5 * time.Second

// A mock implementation of the gRPC Service implementation for testing.
type mockStream struct {
	sendChan chan *Response
	recvChan chan *Request
	ctx      context.Context
}

// newMockStream creates a new mock stream for testing.
func newMockStream() *mockStream {
	// just need a new context specific to this stream,
	// easiest way to get it.
	ctx, _ := context.WithCancel(context.TODO())

	return &mockStream{
		sendChan: make(chan *Response, 100),
		recvChan: make(chan *Request, 100),
		ctx:      ctx,
	}
}

// Send Sends a Response to the mocked client, which
// can then be retrieved by PullSend().
func (m *mockStream) Send(r *Response) error {
	select {
	case m.sendChan <- r:
	case <-time.After(timeOut):
		return errors.New("Timeout on send")
	}

	return nil
}

// PullSend pulls a value out of SendChan.
func (m *mockStream) PullSend() (*Response, error) {
	select {
	case r := <-m.sendChan:
		return r, nil
	case <-time.After(timeOut):
		return nil, errors.New("Timeout on PullSend")
	}
}

// PushRecv Push a Request into the RecvChan.
func (m *mockStream) PushRecv(r *Request) error {
	select {
	case m.recvChan <- r:
	case <-time.After(timeOut):
		return errors.New("Timeout of PushRecv")
	}
	return nil
}

// Recv Receives a Request from the mock client,
// which is given to this mock via PushRecv().
func (m *mockStream) Recv() (*Request, error) {

	select {
	case r := <-m.recvChan:
		if r == nil {
			return r, io.EOF
		}

		return r, nil
	case <-time.After(timeOut):
		return nil, errors.New("Timout on Recv")
	}
}

// Close closes both of the streams used by this mock. Handy for testing.
func (m *mockStream) Close() {
	close(m.recvChan)
	close(m.sendChan)
}

func (m *mockStream) SendHeader(md metadata.MD) error { return nil }
func (m *mockStream) SetTrailer(md metadata.MD)       {}
func (m *mockStream) Context() context.Context        { return m.ctx }
func (m *mockStream) SendMsg(msg interface{}) error   { return nil }
func (m *mockStream) RecvMsg(msg interface{}) error   { return nil }
