/*
  Copyright 2016, Google, Inc.
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

'use strict';

// Needed for the UI
var events = require('events');
var ui_events = new events.EventEmitter();
var ui = require('./ui/ui');
ui.start(processPress, ui_events);


// chalk is used to color cli text output
var chalk = require('chalk');
// grpc interpret protos, server and client tools
var grpc = require('grpc');
// keypress is used to work with cli user input
var keypress = require('keypress');
// variable to enable/disable user input from the cli
var pauseInput = false;

// the game is interactive and will be controlled with keyboard input
// let's setup the keypress listener
keypress(process.stdin);
process.stdin.setRawMode(true);
process.stdin.resume();

process.stdin.on('keypress', function(ch, key) {
    // let's make sure we can exit the programm at any point with ctrl + c
    if (key && key.ctrl && key.name == 'c') {
        process.exit();
    }
    // any other keypress needs to be processed if input isn't paused
    if (!pauseInput) {
        processPress(key.name);
    }
});

// Step 1 - parse and load simonsays.proto, all services, endpoints and messages
// will be available through the proto variable
var proto = grpc.load('simonsays.proto');

// create client connection to SimonSays service
// use serverip and port from environment
var client = new proto.simonsays.SimonSays(
    process.env.server + ":" + process.env.port,
    grpc.credentials.createInsecure()
);

console.log('Welcome to Simon Says!');

// Step 2 - open stream channel to game endpoint
var game = client.game();

// Step 3 - join a game
// to join a game we need a player with a random ID
var player = {
    id: 'Player' + Math.floor(Math.random() * 10000)
};
console.log('Joining with player name ' + player.id);

// we need to send a join request to the game stream channel
// looking at the simonsays.proto this needs to be a message
// of type Request, with the variable join of type Player

game.write({join:player});

// Step 4 - react to events received on the game stream channel
// let's use a callback here
game.on('data', function(event) {
    // Looking at simonsays.proto there're two types of events
    // we need to handle
    // if we receive a lightup event, we need to show the color
    // event.lightup has the value of the COLOR enum we defined in the proto
    if (event.event == 'lightup') {
        // Step 3.1 - handle lightup events
        // send a colored text to the console - use chalk.{color}(text)
        console.log(chalk[event.lightup.toLowerCase()](event.lightup))
        // and send the event to the UI too
        ui_events.emit('lightup', event.lightup);
    }
    else if (event.event == 'turn') {
        // Step 3.2 - handle turn state events
        // Looking at the simonsays.proto there are 5 different states
        // BEGIN - the game starts
        // START_TURN - start accepting user input
        // STOP_TURN - not accepting any user input
        // WIN - game won
        // LOSE - game lost
        switch (event.turn) {
            case 'BEGIN':
                // Join the game
                var msg = 'The game has started \n ' +
                    'To play the game press the keys y, r, g, b for the \n' +
                    ' colors yellow, red, green and blue. \n'
                    'Please wait till it is your turn.';
                console.log(msg);
                ui_events.emit('message', 'The game has started');
                break;
            case 'START_TURN':
                // This player's turn, enable input
                pauseInput = false;
                var msg = 'Your Turn!';
                console.log(msg);
                ui_events.emit('message', msg);
                break;
            case 'STOP_TURN':
                // Others player's turn, disable input
                pauseInput = true;
                var msg = 'It is the other players turn!';
                console.log(msg);
                ui_events.emit('message', msg);
                break;
            case 'WIN':
                var msg = 'You won! :)';
                console.log(msg);
                ui_events.emit('message', msg);
                break;
            case 'LOSE':
                var msg = 'You lost :(';
                console.log(msg);
                ui_events.emit('message', msg);
                break;
        }
    }
});

// Step 5 - if the game stream channel is closed, exit the program
/* REPLACE - handle the closing of the stream channel and exit the program
  Don’t forget to close the channel from client side too ;)
*/
game.on('end', function(event) {
    console.log('Thank you for playing!');
    game.end();
    process.exit();
});


// Processes user input from cli
// Sends the button press to the callback
function processPress(name) {
    //disable input while we process the pressed key
    pauseInput = true;

    var color = '';
    switch (name) {
        case 'r':
            color = 'RED';
            break;
        case 'g':
            color = 'GREEN';
            break;
        case 'y':
            color = 'YELLOW';
            break;
        case 'b':
            color = 'BLUE';
            break;
        default:
            console.log(name + ' is not a valid color input. Try again!')
            break;
    }

    // Step 6 --  send the pressed color to the game stream channel
    if (color != '') {
        game.write({press:color});
    }

    // To avoid accidental double clicks
    setTimeout(function() {
        pauseInput = false
    }, 100);
}