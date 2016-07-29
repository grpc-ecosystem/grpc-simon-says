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

var util = require('util');
var chalk = require('chalk');
var grpc = require('grpc');
var keypress = require('keypress');
var five = require('johnny-five');
var game, board, endState = 'win';
var threshold = process.env.threshold ? parseInt(process.env.threshold) : 10;
var pauseInput = false;
var hadTurn = false;
var button = {
  green: {
    button: null,
    led: null,
    prev: -1
  },
  red: {
    button: null,
    led: null,
    prev: -1
  },
  yellow: {
    button: null,
    led: null,
    prev: -1
  },
  blue: {
    button: null,
    led: null,
    prev: -1
  },
};

//Read in hardware config file
var config = require('./' + process.env.config);

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

//Initiate the Arduino Board with Johnny-Five
board = new five.Board({repl: false});
board.on('ready', function() {

  button.red.led = new five.Led(config.buttons.red.led);
  button.blue.led = new five.Led(config.buttons.blue.led);
  button.yellow.led = new five.Led(config.buttons.yellow.led);
  button.green.led = new five.Led(config.buttons.green.led);

  if (config.type == 'analog') {
    button.red.button = new five.Sensor({pin: config.buttons.red.button});
    button.blue.button = new five.Sensor({pin: config.buttons.blue.button});
    button.yellow.button = new five.Sensor({pin: config.buttons.yellow.button});
    button.green.button = new five.Sensor({pin: config.buttons.green.button});
  } else if (config.type == 'digital') {
    button.red.button = new five.Button({
      pin: config.buttons.red.button.pin,
      invert: config.buttons.red.button.invert,
      isPullup: config.buttons.red.button.isPullup
    });
    button.blue.button = new five.Button({
      pin: config.buttons.blue.button.pin,
      invert: config.buttons.blue.button.invert,
      isPullup: config.buttons.blue.button.isPullup
    });
    button.yellow.button = new five.Button({
      pin: config.buttons.yellow.button.pin,
      invert: config.buttons.yellow.button.invert,
      isPullup: config.buttons.yellow.button.isPullup
    });
    button.green.button = new five.Button({
      pin: config.buttons.green.button.pin,
      invert: config.buttons.green.button.invert,
      isPullup: config.buttons.green.button.isPullup
    });
  } else {
    console.log('Invalid Config');
    endProgram();
  }
  start();
});

function newGame() {

  // gameStatus is a server stream that registers the player to the server
  // and gets a stream of game events (BEGIN, START_TURN, END_TURN, WIN, LOSE)
  game = client.game();

  //Join Game
  var event = {
    join: player
  };
  game.write(event);

  //Arduino Button Presses
  function buttonHandler(value, color) {
    //Check that there is a legit previous value
    //And that the touch is over the trigger threshold
    if (button[color].prev > -1 && value > button[color].prev + threshold) {
      console.log(chalk.magenta(color + ' triggered'));
      if (!pauseInput) {
        processPress(color);
      }
    }
    button[color].prev = value > 1 ? value : -1;
  }
  if (config.type == 'analog') {
    button.red.button.on('data', function() {
      buttonHandler(this.value, 'red');
    });
    button.blue.button.on('data', function() {
      buttonHandler(this.value, 'blue');
    });
    button.yellow.button.on('data', function() {
      buttonHandler(this.value, 'yellow');
    });
    button.green.button.on('data', function() {
      buttonHandler(this.value, 'green');
    });
  }else {
    button.red.button.on('press', function() {
      processPress('red');
    });
    button.blue.button.on('press', function() {
      processPress('blue');
    });
    button.yellow.button.on('press', function() {
      processPress('yellow');
    });
    button.green.button.on('press', function() {
      processPress('green');
    });
  }

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
          joinAnimation();
          break;

        case 'START_TURN':
          // Start of turn, enable input
          hadTurn = true;
          console.log('Your Turn:');
          setTimeout(function() {
            turnOnLeds({red: true, green: true, blue: true, yellow: true});
            turnOffLeds(1000, null);
            pauseInput = false;
          }, 1000);
          break;

        case 'STOP_TURN':
          // Player's turn is over, disable input
          if (hadTurn) {
            console.log("Nice Job! Now it's the other player's turn");
          }else {
            console.log("It's the other player's turn first");
          }
          pauseInput = true;
          break;

        case 'WIN':
          console.log('You Win!');
          endState = 'WIN';
          break;

        case 'LOSE':
          console.log('You Lost!');
          endState = 'LOSE';
          turnOnLeds({red: true});
          turnOffLeds(500, function() {
            turnOnLeds({red: true});
            turnOffLeds(500, function() {
              turnOnLeds({red: true});
              turnOffLeds(500, null);
            });
          });
          break;
      }
    }
  });
  game.on('end', function() {
    //Game stream is over, shut down the game
    var color = {red: true};
    if (endState == 'WIN')
      color = {green: true};
    setTimeout(function() {
      turnOnLeds(color);
      turnOffLeds(2000, function() {
        turnOnLeds(color);
        turnOffLeds(2000, function() {
          turnOnLeds(color);
          console.log('Thanks for Playing');
          endProgram();
        });
      });
    }, 1000);
  });

}

// Light up the color (and print to console)
function lightUp(color) {
  switch (color) {
    case 'RED':
      turnOnLeds({red: true});
      console.log(chalk.red(color));
      break;
    case 'GREEN':
      turnOnLeds({green: true});
      console.log(chalk.green(color));
      break;
    case 'YELLOW':
      turnOnLeds({yellow: true});
      console.log(chalk.yellow(color));
      break;
    case 'BLUE':
      turnOnLeds({blue: true});
      console.log(chalk.blue(color));
      break;
  }
  turnOffLeds(200, null);
}

// Processes a input
// Sends the button press to the callback
function processPress(name) {
  //Disable Input
  pauseInput = true;

  var color = 'RED';
  switch (name) {
    case 'a':
    case 'red':
      color = 'RED';
      break;
    case 's':
    case 'green':
      color = 'GREEN';
      break;
    case 'd':
    case 'yellow':
      color = 'YELLOW';
      break;
    case 'f':
    case 'blue':
      color = 'BLUE';
      break;
  }

  game.write({press: color});

  // Avoid accidental double clicks
  setTimeout(function() {pauseInput = false}, 200);
}

// Turns off all LEDs after a set delay
function turnOffLeds(delay, callback) {
  // Check if there was an actual function passed in
  // If not, make callback an empty function
  if (!(util.isFunction(callback))) {
    callback = function() {};
  }
  setTimeout(function() {
    button.red.led.off();
    button.blue.led.off();
    button.yellow.led.off();
    button.green.led.off();
    callback();
  }, delay);
}

// Turns on all the LEDs
function turnOnLeds(leds) {
  if (leds.red == true) {
    button.red.led.on();
  }
  if (leds.blue == true) {
    button.blue.led.on();
  }
  if (leds.yellow == true) {
    button.yellow.led.on();
  }
  if (leds.green == true) {
    button.green.led.on();
  }
}

function start() {
  console.log('Welcome to Simon Says!');

  // Some beautiful callback code for the boot animation
  turnOnLeds({red: true});
  turnOffLeds(200, function() {
    turnOnLeds({blue: true});
    turnOffLeds(200, function() {
      turnOnLeds({yellow: true});
      turnOffLeds(200, function() {
        turnOnLeds({green: true});
        turnOffLeds(200, function() {
          turnOnLeds({red: true, blue: true, yellow: true, green: true});
          turnOffLeds(100, null);
          // Start the actual game
          newGame();
        });
      });
    });
  });
}

function joinAnimation() {
  console.log('Game Started');
  turnOnLeds({yellow: true, red: true});
  turnOffLeds(200, function() {
    turnOnLeds({green: true, blue: true});
    turnOffLeds(200, function() {
      turnOnLeds({red: true, blue: true, yellow: true, green: true});
      turnOffLeds(100, null);
    });
  });
}

function endProgram() {
  console.log('Goodbye');
  turnOffLeds(10, null);
  setTimeout(function() {
    game.cancel();
    process.exit();
  }, 100);
}
