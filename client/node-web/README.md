# Web Client for gRPC Simon

This is an implementation of the gRPC Simon client in Node.js that uses the command line 
or a web browser as the input method.
It uses Socket.io to communicate with the browser over websockets.

## How to deploy

- Install dependencies
     - `make install`
- Run:
     - The server port is optional and defaults to 50051.
     - The local port is optional and defaults to 8080.
     - `make run SERVERIP=<server-ip-here> SERVERPORT=<port-ip-here> LOCALPORT=<port-ip-here>`
- Open in browser:
     - Replace 8080 with your `LOCALPORT`
     - `localhost:8080`

Notes:
- This is not an official Google product
- Only tested on OSX and Linux (Ubuntu)
  - If you test on Windows, let us know! Pull requests welcome.