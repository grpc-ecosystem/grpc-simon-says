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

/*
Server binary that actually runs the gRPC server
*/
package main

import (
	"log"
	"net"
	"os"

	"github.com/grpc-simonsays/simonsays-server/simonsays"
	"google.golang.org/grpc"
)

const (
	port         = "PORT"
	redisAddress = "REDIS_ADDRESS"
)

// Create a Server instance and fire it up!
func main() {
	port := os.Getenv(port)
	// default for port
	if port == "" {
		port = "50051"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("[Error][Server] Could not listen on port %v. %v", port, err)
	}
	defer lis.Close()

	s := grpc.NewServer()

	simon, err := simonsays.NewSimonSays(os.Getenv(redisAddress))
	if err != nil {
		log.Fatalf("[Error][Server] Could not connect to redis: %v.", err)
	}
	defer simon.Close()

	simonsays.RegisterSimonSaysServer(s, simon)

	log.Printf("[Info][Server] Starting server on port %v", port)
	log.Printf("[Info][Server] The server has been stopped: %v", s.Serve(lis))
}
