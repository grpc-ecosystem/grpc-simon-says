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

// Package logger is logging that is specific to this application
// Need to track the game, and the player, so we can see
// everything that is going on.
package logger

import (
	"bytes"
	"fmt"
	"log"
	"sync"

	"golang.org/x/net/context"
)

var (
	data = map[context.Context]map[string]string{}
	// this is to track keys in the order they are set
	keys = map[context.Context][]string{}
	lock sync.RWMutex
)

// Set set a value for this context
func Set(ctx context.Context, key, value string) {
	lock.Lock()
	defer lock.Unlock()

	if _, ok := data[ctx]; !ok {
		data[ctx] = map[string]string{}
	}

	data[ctx][key] = value
	keys[ctx] = append(keys[ctx], key)

}

// Clear clears the data for this context. Very important
// to do at the end of a request, to ensure no memory leaks
func Clear(ctx context.Context) {
	lock.Lock()
	defer lock.Unlock()

	delete(data, ctx)
	delete(keys, ctx)
}

// Info informational level logging.
func Info(ctx context.Context, category, msg string, args ...interface{}) {
	printf(ctx, "Info", category, msg, args...)
}

// Error error level logging.
func Error(ctx context.Context, category, msg string, args ...interface{}) {
	printf(ctx, "Error", category, msg, args...)
}

func printf(ctx context.Context, level, category, msg string, args ...interface{}) {
	lock.RLock()
	defer lock.RUnlock()

	buf := new(bytes.Buffer)
	_, err := buf.WriteString(fmt.Sprintf("[%v][%v]", level, category))

	if err != nil {
		log.Println("[Error][Logging] Could not write level and category to buffer.")
	}

	for _, k := range keys[ctx] {
		_, err = buf.WriteString(fmt.Sprintf("[%v: %v]", k, data[ctx][k]))
		if err != nil {
			log.Println("[Error][Logging] Could not write key, value to the buffer")
		}
	}

	_, err = buf.WriteString(fmt.Sprintf(" "+msg, args...))

	if err != nil {
		log.Println("[Error][Logging] Could not write message to buffer.")
	}

	log.Println(buf.String())
}
