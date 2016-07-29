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

var http = require('http');
var fs = require('fs');
var file = fs.readFileSync(__dirname + '/index.html', 'utf8');
var port = 8080;

module.exports = {
  start: function(clickHandler, ui_events) {

    var handleRequest = function(request, response) {
      response.writeHead(200, {"Content-Type": "text/html"});
      response.end(file);
    };

    var server = http.createServer(handleRequest);
    var io = require('socket.io')(server);

    io.on('connection', function(socket) {
      socket.on('click', function(data) {
        clickHandler(data);
      });
    });

    ui_events.on('lightup', function(color) {
      io.emit('lightup', color);
    });

    ui_events.on('message', function(data) {
      io.emit('message', data);
    });

    server.listen(port, function() {
      console.log('UI Server: http://localhost:%s', port);
    });
  },
};
