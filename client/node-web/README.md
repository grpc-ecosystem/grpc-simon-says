# Web Client for gRPC Simon

This is an implementation of the gRPC Simon client in Node.js that uses a web browser as the input method.
It uses Socket.io to communicate with the browser over websockets.

## How to deploy

- Install dependencies
     - `make install`
- Run:
     - `make run SERVERIP=<server-ip-here>`
- Open in browser:
     - `localhost:8080`

Notes:
- This is not an official Google product
- Only tested on OSX and Linux (Ubuntu)
  - If you test on Windows, let us know! Pull requests welcome.
