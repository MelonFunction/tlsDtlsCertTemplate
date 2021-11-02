# Simple setup of DTLS and TLS. 
Mostly copied from the `pion/dtls` examples.

Run `./makecerts.sh` and then start the server with `go run ./server`. 
Connect a client with `go run ./client`.
The server simply echos the client stdin.