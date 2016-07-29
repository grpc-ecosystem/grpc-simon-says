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

var util = require('util');
var chalk = require('chalk');
var grpc = require('grpc');
var keypress = require('keypress');
var game, board;
var pauseInput = false;
var hadTurn = false;

//Create a random PlayerID
var player = {
  id: 'Player' + Math.floor(Math.random() * 10000)
};

console.log(player.id);

//Connect to the gRPC server
var proto = grpc.load('simonsays.proto');
var client = new proto.simonsays.SimonSays(
  process.env.server + ':' + process.env.port,
  grpc.credentials.createInsecure()
);

startKeyboardCapture();
console.log('Welcome to Simon Says!');
newGame();

// Start the UI
ui.start(processPress, ui_events);

function newGame() {

  // gameStatus is a server stream that registers the player to the server
  // and gets a stream of game events (BEGIN, START_TURN, END_TURN, WIN, LOSE)
  game = client.game();

  //Join Game
  var event = {
    join: player
  };
  game.write(event);


  //Keyboard Input
  process.stdin.on('keypress', function(ch, key) {
    if (key && key.ctrl && key.name == 'c') {
      process.exit();
    }
    if (!pauseInput) {
      processPress(key.name);
    }
  });

  //Get Game Events
  game.on('data', function(event) {

    //Type of event is Color
    if (event.event == 'lightup') {
      lightUp(event.lightup);
    }
    //Type of event is Game Sate
    else if (event.event == 'turn') {
      switch (event.turn) {

        case 'BEGIN':
          // Join the game
          var str = 'The game has started';
          ui_events.emit('message', str);
          console.log(str);
          break;

        case 'START_TURN':
          // Start of turn, enable input
          hadTurn = true;
          pauseInput = false;
          var str = 'Your Turn!';
          ui_events.emit('message', str);
          console.log(str);
          break;

        case 'STOP_TURN':
          // Player's turn is over, disable input
          if (hadTurn) {
            var str = "Nice Job! Now it's the other player's turn";
            ui_events.emit('message', str);
            console.log(str);
          }else {
            var str = "It's the other player's turn first";
            ui_events.emit('message', str);
            console.log(str);
          }
          pauseInput = true;
          break;

        case 'WIN':
          var str = 'You Win!';
          ui_events.emit('message', str);
          console.log(str);
          break;

        case 'LOSE':
          var str = 'You Lost!';
          ui_events.emit('message', str);
          console.log(str);
          break;
      }
    }
  });

  game.on('end', function() {
    var str = 'Thanks for Playing';
    ui_events.emit('message', str);
    console.log(str);
    game.cancel();
    process.exit();
  });

}

// Light up the color (and print to console)
function lightUp(color) {

  // Emit the event so the UI can capture it
  ui_events.emit('lightup', color);

  switch (color) {
    case 'RED':
      console.log(chalk.red(color));
      break;
    case 'GREEN':
      console.log(chalk.green(color));
      break;
    case 'YELLOW':
      console.log(chalk.yellow(color));
      break;
    case 'BLUE':
      console.log(chalk.blue(color));
      break;
  }
}

// Processes a input
// Sends the button press to the callback
function processPress(name) {
  //Disable Input
  pauseInput = true;

  var color = 'RED';
  switch (name) {
    case 'a':
    case 'RED':
      color = 'RED';
      break;
    case 's':
    case 'GREEN':
      color = 'GREEN';
      break;
    case 'd':
    case 'YELLOW':
      color = 'YELLOW';
      break;
    case 'f':
    case 'BLUE':
      color = 'BLUE';
      break;
  }

  game.write({press: color});

  // Avoid accidental double clicks
  setTimeout(function() {pauseInput = false}, 200);
}

function startKeyboardCapture() {
  //Capture the raw input to emulate button presses
  keypress(process.stdin);
  process.stdin.setRawMode(true);
  process.stdin.resume();
}
