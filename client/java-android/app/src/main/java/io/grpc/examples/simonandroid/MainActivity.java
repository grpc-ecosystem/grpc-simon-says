/*
 *    Copyright 2016 Google Inc.
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package io.grpc.examples.simonandroid;

import android.app.AlertDialog;
import android.app.Dialog;
import android.content.DialogInterface;
import android.os.Bundle;
import android.support.v7.app.AppCompatActivity;
import android.util.Log;
import android.view.LayoutInflater;
import android.view.View;
import android.view.animation.AlphaAnimation;
import android.view.animation.Animation;
import android.widget.Button;
import android.widget.EditText;
import android.widget.Toast;

import java.util.Random;

import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.examples.simonsays.nano.SimonSaysGrpc;
import io.grpc.examples.simonsays.nano.Request;
import io.grpc.examples.simonsays.nano.Response;
import io.grpc.examples.simonsays.nano.SimonSays;
import io.grpc.stub.StreamObserver;

public class MainActivity extends AppCompatActivity {

    private static final int PORT = 50051;
    private static final String LOG_TAG = MainActivity.class.getName();
    private ManagedChannel channel;
    private StreamObserver<Request> requests;
    private SimonSaysGrpc.SimonSaysStub asyncStub;
    private String playerId;
    private Animation buttonAnimation;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_main);
        createPlayerId();
        createButtonAnimation();
        createConnectDialog().show();
    }

    private Dialog createConnectDialog() {
        AlertDialog.Builder builder = new AlertDialog.Builder(this);
        LayoutInflater inflater = getLayoutInflater();
        // Inflate and set the layout for the dialog
        // Pass null as the parent view because its going in the dialog layout
        final View dialogView = inflater.inflate(R.layout.signin, null);
        builder.setTitle(R.string.sign_in)
        .setView(dialogView)
                // Add action buttons
                .setPositiveButton(R.string.start_game, new DialogInterface.OnClickListener() {
                    @Override
                    public void onClick(DialogInterface dialog, int id) {
                        EditText address = (EditText)dialogView.findViewById(R.id.serverAddress);
                        createClient(address.getText().toString());
                        requests = startGame();
                        joinGame(requests);
                    }
                })
                .setNegativeButton(R.string.cancel, new DialogInterface.OnClickListener() {
                    public void onClick(DialogInterface dialog, int id) {
                        //can't do anything, bail.
                        System.exit(0);
                    }
                });

        return builder.create();
    }

    /*
    * Join a game
    */
    private void joinGame(StreamObserver<Request> requests) {
        Log.i(LOG_TAG, "Joining a game");
        Request.Player player = new Request.Player();
        player.id = playerId;
        Request request = new Request().setJoin(player);
        sendRequest(request);
    }

    /*
    * Create the animation to use on all buttons
    */
    private void createButtonAnimation() {
        buttonAnimation = new AlphaAnimation(0.2f, 1f);
        buttonAnimation.setDuration(500);
    }

    /*
    * Start a game!
    */
    private StreamObserver<Request> startGame() {
        Log.i(getClass().getName(), "Starting up a game...");
        return asyncStub.game(new StreamObserver<Response>() {
            @Override
            public void onNext(Response value) {
                receiveResponse(value);
            }

            @Override
            public void onError(Throwable t) {
                Log.e(LOG_TAG, "Uh Oh. Error.", t);
            }

            @Override
            public void onCompleted() {
                Log.i(LOG_TAG, "Disconnected.");
            }
        });
    }

    /*
    * Button click handler: blue
    */
    public void onBlue(View view) {
        Log.i(LOG_TAG, "Pressed Blue!");
        sendPress(SimonSays.BLUE);
    }

    /*
    * Button click handler: red
    */
    public void onRed(View view) {
        Log.i(LOG_TAG, "Pressed Red");
        sendPress(SimonSays.RED);
    }

    /*
    * Button click handler: green
    */
    public void onGreen(View view) {
        Log.i(LOG_TAG, "Pressed Green!");
        sendPress(SimonSays.GREEN);
    }

    /*
    * Button click handler: yellow
    */
    public void onYellow(View view) {
        Log.i(LOG_TAG, "Pressed Yellow!");
        sendPress(SimonSays.YELLOW);
    }

    /*
    * Creating the Player's unique Id
    */
    private void createPlayerId() {
        playerId = "Android" + (new Random().nextInt(99999));
        Log.i(getClass().getName(), "My playerId is: " + playerId);
    }

    /*
    * Create the gRPC Client
    */
    private void createClient(String host) {
        Log.i(this.getClass().getName(), "Creating gRPC Client , at address: " + host);
        channel = ManagedChannelBuilder.forAddress(host, PORT)
                .usePlaintext(true)
                .build();

        asyncStub = SimonSaysGrpc.newStub(channel);
    }

    /*
    * sends a request to the gRPC server
    * @stream push stream
    */
    private void sendRequest(Request request) {
        Log.i(getClass().getName(), "Sending Request: " + request.toString());
        requests.onNext(request);
    }

    /*
    * Send a button press
    */
    private void sendPress(int color) {
        Request request = new Request();
        request.setPress(color);
        sendRequest(request);
    }

    /*
    * Wrapper for receiving responses
    */
    private void receiveResponse(Response value) {
        Log.i(LOG_TAG, "Response Received: " + value.toString());

        final Response finalValue = value;
        runOnUiThread(new Runnable() {
            @Override
            public void run() {
                switch (finalValue.getEventCase()) {
                    case Response.TURN_FIELD_NUMBER: {
                        handleTurn(finalValue.getTurn());
                        break;
                    }

                    case Response.LIGHTUP_FIELD_NUMBER: {
                        handleLightup(finalValue.getLightup());
                        break;
                    }
                }
            }
        });
    }

    /*
    * Handle lightup events
    */
    private void handleLightup(int lightup) {
        switch (lightup) {
            case SimonSays.RED: {
                Log.i(LOG_TAG, "Lightup: RED");
                Button button = (Button) findViewById(R.id.button_red);
                button.startAnimation(buttonAnimation);
                break;
            }

            case SimonSays.YELLOW: {
                Log.i(LOG_TAG, "Lightup: YELLOW");
                Button button = (Button) findViewById(R.id.button_yellow);
                button.startAnimation(buttonAnimation);
                break;
            }

            case SimonSays.GREEN: {
                Log.i(LOG_TAG, "Lightup: GREEN");
                Button button = (Button) findViewById(R.id.button_green);
                button.startAnimation(buttonAnimation);
                break;
            }

            case SimonSays.BLUE: {
                Log.i(LOG_TAG, "Lightup: BLUE");
                Button button = (Button) findViewById(R.id.button_blue);
                button.startAnimation(buttonAnimation);
                break;
            }
        }
    }

    /*
    * Handle turn events
    */
    private void handleTurn(int turn) {
        switch (turn) {
            case Response.BEGIN: {
                Log.i(LOG_TAG, "It's my turn!");
                Toast.makeText(getApplicationContext(), "Welcome to Simon Says", Toast.LENGTH_SHORT).show();
                break;
            }

            case Response.START_TURN: {
                Log.i(LOG_TAG, "Starting turn");
                Toast.makeText(getApplicationContext(), "It's Your Turn.", Toast.LENGTH_SHORT).show();
                break;
            }

            case Response.STOP_TURN: {
                Log.i(LOG_TAG, "Stopping Turn");
                Toast.makeText(getApplicationContext(), "Other Players Turn...", Toast.LENGTH_SHORT).show();
                break;
            }

            case Response.WIN: {
                Log.i(LOG_TAG, "WON!");
                AlertDialog.Builder builder = new AlertDialog.Builder(this);
                builder.setTitle("YOU WON :)").setPositiveButton("CLOSE", new DialogInterface.OnClickListener() {
                    @Override
                    public void onClick(DialogInterface dialog, int which) {
                        requests.onCompleted();
                        System.exit(0);
                    }
                });
                builder.create().show();
                break;
            }

            case Response.LOSE: {
                Log.i(LOG_TAG, "LOST");
                AlertDialog.Builder builder = new AlertDialog.Builder(this);
                builder.setTitle("YOU LOST :(").setPositiveButton("CLOSE", new DialogInterface.OnClickListener() {
                    @Override
                    public void onClick(DialogInterface dialog, int which) {
                        requests.onCompleted();
                        System.exit(0);
                    }
                });
                builder.create().show();
                break;
            }
        }
    }
}
