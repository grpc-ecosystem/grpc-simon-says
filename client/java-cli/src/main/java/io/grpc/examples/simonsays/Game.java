/*
 * Copyright 2016 Google Inc. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package io.grpc.examples.simonsays;

import java.util.logging.Level;
import java.util.logging.Logger;

import io.grpc.examples.simonsays.Color;
import io.grpc.examples.simonsays.Request;
import io.grpc.examples.simonsays.Response;
import io.grpc.examples.simonsays.Response.State;
import io.grpc.examples.simonsays.SimonSaysGrpc;
import io.grpc.stub.StreamObserver;

/**
 * Object encapsulating state for a Simon Says Game.
 */
public class Game {

    private final static Logger Log = Logger.getLogger(Game.class.getName());

    private final ConsoleWindow window;
    private final SimonSaysGrpc.SimonSaysStub asyncStub;

    private StreamObserver<Request> streamObserver;
    private boolean started;

    /**
     * Constructor for Game.
     */
    public Game(ConsoleWindow window,
            SimonSaysGrpc.SimonSaysStub asyncStub) {
        this.window = window;
        this.asyncStub = asyncStub;

        setupStreamObserver();
    }

    /**
     * Send a join game request.
     */
    public void join(String playerId) {
        Log.log(Level.INFO, "Joining a game...");

        Request.Player player = Request.Player.newBuilder().setId(playerId).build();
        Request request = Request.newBuilder().setJoin(player).build();
        streamObserver.onNext(request);
    }

    /**
     * Send a color.
     */
    public void sendColor(Color color) {
        Log.log(Level.INFO, "Sending color: " + color);

        Request request = Request.newBuilder().setPress(color).build();
        streamObserver.onNext(request);
    }

    /**
     * Return the startedGame boolean.
     */
    public boolean isStarted() {
        return started;
    }

    /**
     * Setup stream observer.
     */
    private void setupStreamObserver() {
        Log.log(Level.INFO, "Setting up for game...");

        streamObserver = asyncStub.game(new StreamObserver<Response>() {
            @Override
            public void onNext(Response value) {
                handleResponse(value);
            }

            @Override
            public void onError(Throwable t) {
                Log.log(Level.SEVERE, "Error", t);
            }

            @Override
            public void onCompleted() {
                Log.log(Level.INFO, "Disconnected");
                System.exit(0);
            }
        });
    }

    /**
     * Handle response value which can either by lightup or turn.
     */
    private void handleResponse(Response value) {
        Log.log(Level.INFO, "Response received: " + value.toString());

        switch (value.getEventCase().getNumber()) {
            case Response.LIGHTUP_FIELD_NUMBER: {
                handleLightup(value.getLightup());
                break;
            }
            case Response.TURN_FIELD_NUMBER: {
                handleTurn(value.getTurn());
                break;
            }
        }
    }

    /**
     * Handle lightup event.
     */
    private void handleLightup(Color color) {
        window.printColoredSimonCube(color);
    }

    /**
     * Handle turn event.
     */
    private void handleTurn(State state) {
        switch (state) {
            case BEGIN: {
                started = true;
                Log.log(Level.INFO, "Game is starting!");
                window.println("Game is starting...");
                window.printBlankSimonCube();
                break;
            }

            case START_TURN: {
                Log.log(Level.INFO, "Starting turn");
                window.println("It's your turn. Press r,g,b,y to choose Red, Green, Blue, Yellow");
                break;
            }

            case STOP_TURN: {
                Log.log(Level.INFO, "Stopping Turn");
                window.println("It's your opponent's turn...");
                break;
            }

            case WIN: {
                Log.log(Level.INFO, "WON!");
                window.println("You WON!");
                started = false;
                break;
            }

            case LOSE: {
                Log.log(Level.INFO, "LOST");
                window.println("You LOST!");
                started = false;
                break;
            }

            case UNRECOGNIZED:
            default:
                break;
        }
    }
}