/**
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

import java.io.IOException;
import java.util.logging.Level;
import java.util.logging.LogManager;
import java.util.logging.Logger;

import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.examples.simonsays.Color;

/**
 * Entry point to the game.
 */
public class Main {

    private final static Logger Log = Logger.getLogger(Main.class.getName());

    private final ConsoleInputReader reader;
    private final ConsoleWindow window;

    public static void main(String[] args) throws InterruptedException, IOException {
        Main main = new Main();
        main.run();
    }

    /**
     * Default constructor.
     */
    public Main() {
        // Create a reader and a window.
        reader = new ConsoleInputReader();
        window = new ConsoleWindow();

        // Disable logging by default.
        LogManager.getLogManager().reset();
    }

    /**
     * Main run loop.
     */
    private void run() throws InterruptedException, IOException {
        window.println("Welcome to Simon Says!");
        window.println("First, some setup...");

        // Create a game.
        Game game = createGame();

        // Read player id.
        String playerId = reader.readPlayerName();

        // Either create a game or join an already waiting player.
        game.join(playerId);

        // Block until another player joins.
        window.println("Waiting for another player to join...");
        while (!game.isStarted()) {
            Thread.sleep(1000);
        }

        // Main game loop.
        while (true) {
            Color color = reader.readColor();
            if (color == Color.UNRECOGNIZED) {
                window.println("Unrecognized color, valid input [r,g,y,b]...");
            } else {
                game.sendColor(color);
            }
        }
    }

    /**
     * Create a game. In order to create a game, you need:
     * 
     * Read server ip and server port (using reader) and create a ManagedChannel.
     * Create an async stub using the created ManagedChannel.
     */
    private Game createGame() {

        ManagedChannel channel = createManagedChannel();

        SimonSaysGrpc.SimonSaysStub asyncStub = createAsyncStub(channel);;

        return new Game(window, asyncStub);
    }

    /**
     * Read server ip and server port (using reader) and create a ManagedChannel.
     */
    private ManagedChannel createManagedChannel() {
        String serverIp = reader.readServerIp();
        int serverPort = reader.readServerPort();

        Log.log(Level.INFO, "Creating a managed channel for address: " 
                + (serverIp + ":" + serverPort));

        ManagedChannel channel = ManagedChannelBuilder.forAddress(serverIp, serverPort)
                .usePlaintext(true)
                .build();

        return channel;
    }

    /**
     * Create an async stub using the created ManagedChannel.
     */
    private SimonSaysGrpc.SimonSaysStub createAsyncStub(ManagedChannel channel) {
        Log.log(Level.INFO, "Creating an async SimonSaysGrpc stub");

        return SimonSaysGrpc.newStub(channel);
    }
}
