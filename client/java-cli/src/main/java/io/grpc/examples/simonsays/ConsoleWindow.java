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

import org.fusesource.jansi.AnsiConsole;

import io.grpc.examples.simonsays.Color;

/**
 * Class that encapsulates a console window.
 */
public class ConsoleWindow {

    private static final String ANSI_RED = "\u001B[31m";
    private static final String ANSI_GREEN = "\u001B[32m";
    private static final String ANSI_BLUE = "\u001B[34m";
    private static final String ANSI_YELLOW = "\u001B[33m";
    private static final String ANSI_RESET = "\u001B[0m";

    /**
     * Print text.
     */
    public void println(String text) {
        AnsiConsole.out.println(text);
    }

    /**
     * Print Simon Says Cube with defined color.
     */
    public void printColoredSimonCube(Color color) {
        clearScreen();

        switch (color) {
            case RED:
                printRed();
                break;
            case GREEN:
                printGreen();
                break;
            case BLUE:
                printBlue();
                break;
            case YELLOW:
                printYellow();
                break;
            default:
                break;
        }

        // Sleep, to show the colored cube briefly.
        try {
            Thread.sleep(750);
        } catch (InterruptedException e) {
            e.printStackTrace();
        }

        // Then, clear the screen and show the blank cube.
        clearScreen();
        printBlankSimonCube();
    }

    /**
     * Print a blank Simon Says cube.
     */
    public void printBlankSimonCube() {
        AnsiConsole.out.println("+----+----+");
        AnsiConsole.out.println("|    |    |");
        AnsiConsole.out.println("|    |    |");
        AnsiConsole.out.println("+----+----+");
        AnsiConsole.out.println("|    |    |");
        AnsiConsole.out.println("|    |    |");
        AnsiConsole.out.println("+----+----+");
    }

    private static void printRed() {
        AnsiConsole.out.println(ANSI_RED + "+----+" + ANSI_RESET + "----+");
        AnsiConsole.out.println(ANSI_RED + "|RRRR|" + ANSI_RESET + "    |");
        AnsiConsole.out.println(ANSI_RED + "|RRRR|" + ANSI_RESET + "    |");
        AnsiConsole.out.println(ANSI_RED + "+----+" + ANSI_RESET + "----+");
        AnsiConsole.out.println("|    |    |");
        AnsiConsole.out.println("|    |    |");
        AnsiConsole.out.println("+----+----+");
    }

    private void printGreen() {
        AnsiConsole.out.println("+----" + ANSI_GREEN + "+----+" + ANSI_RESET);
        AnsiConsole.out.println("|    " + ANSI_GREEN + "|GGGG|" + ANSI_RESET);
        AnsiConsole.out.println("|    " + ANSI_GREEN + "|GGGG|" + ANSI_RESET);
        AnsiConsole.out.println("+----" + ANSI_GREEN + "+----+" + ANSI_RESET);
        AnsiConsole.out.println("|    |    |");
        AnsiConsole.out.println("|    |    |");
        AnsiConsole.out.println("+----+----+");
    }

    private void printBlue() {
        AnsiConsole.out.println("+----+----+");
        AnsiConsole.out.println("|    |    |");
        AnsiConsole.out.println("|    |    |");
        AnsiConsole.out.println(ANSI_BLUE + "+----+" + ANSI_RESET + "----+");
        AnsiConsole.out.println(ANSI_BLUE + "|BBBB|" + ANSI_RESET + "    |");
        AnsiConsole.out.println(ANSI_BLUE + "|BBBB|" + ANSI_RESET + "    |");
        AnsiConsole.out.println(ANSI_BLUE + "+----+" + ANSI_RESET + "----+");
    }

    private void printYellow() {
        AnsiConsole.out.println("+----+----+");
        AnsiConsole.out.println("|    |    |");
        AnsiConsole.out.println("|    |    |");
        AnsiConsole.out.println("+----" + ANSI_YELLOW + "+----+" + ANSI_RESET);
        AnsiConsole.out.println("|    " + ANSI_YELLOW + "|YYYY|" + ANSI_RESET);
        AnsiConsole.out.println("|    " + ANSI_YELLOW + "|YYYY|" + ANSI_RESET);
        AnsiConsole.out.println("+----" + ANSI_YELLOW + "+----+" + ANSI_RESET);
    }

    private void clearScreen() {
        AnsiConsole.out.print(String.format("\033[2J"));
    }
}
