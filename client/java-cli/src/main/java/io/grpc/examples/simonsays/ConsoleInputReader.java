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

import java.util.Random;
import java.util.Scanner;

import io.grpc.examples.simonsays.Color;

/**
 * Utility class to read input from console.
 */
public class ConsoleInputReader {

    private static final String DEFAULT_PLAYER_NAME = "JavaPlayer";
    private static final String DEFAULT_SERVER_IP = "127.0.0.1";
    private static final int DEFAULT_SERVER_PORT = 50051;
    private static final String INPUT_FORMAT = "%s [%s]: ";

    private final Scanner scanner;

    public ConsoleInputReader() {
        scanner = new Scanner(System.in);
    }

    public Color readColor() {
        String line = scanner.next();
        Color color = doReadColor(line);
        scanner.reset();
        return color;
    }

    public String readPlayerName() {
        return readString("Player name", DEFAULT_PLAYER_NAME + "-" + new Random().nextInt(10000));
    }

    public String readServerIp() {
        return readString("Game Server IP", DEFAULT_SERVER_IP);
    }

    public int readServerPort() {
        return readInt("Game Server Port", DEFAULT_SERVER_PORT);
    }

    private String readString(String question, String defaultValue) {
        String line = readLine(question, defaultValue);
        return isNullOrEmpty(line) ? defaultValue : line;
    }

    private int readInt(String question, int defaultValue) {
        try {
            String line = readLine(question, defaultValue);
            return Integer.parseInt(line);
        } catch (Exception e) {
            return defaultValue;
        }
    }
    
    private String readLine(String question, Object defaultValue) {
        System.out.print(getFormattedQuestion(question, defaultValue));
        String line =  scanner.nextLine();
        return line;
    }

    private static String getFormattedQuestion(String question, Object defaultValue) {
        return String.format(INPUT_FORMAT, question, defaultValue);
    }

    private static boolean isNullOrEmpty(String str) {
        return str == null || str.isEmpty();
    }

    private static Color doReadColor(String line) {
        if (line.equalsIgnoreCase("b")) {
            return Color.BLUE;
        }
        if (line.equalsIgnoreCase("g")) {
            return Color.GREEN;
        }
        if (line.equalsIgnoreCase("r")) {
            return Color.RED;
        }
        if (line.equalsIgnoreCase("y")) {
            return Color.YELLOW;
        }
        return Color.UNRECOGNIZED;
    }
}
