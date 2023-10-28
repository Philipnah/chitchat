# chitchat

gRPC based logical time distributed system, implemented in go.

# User manual — How to use the program

1. In the root directory (the "chitchat" folder), start the server first, with the following command:

```
go run server/server.go
```

2. Then start a client with the following command:

```
go run client/client.go
```

The client will clear the terminal once started, then display the "joining" message, on all connected clients. You may start many clients.

3. Once a client has been started, you can start writing messages, pressing "Enter" to send each message. These messages will appear for all participants, just like in a chatroom.

4. To exit the chat, interrupt the terminal process by pressing `ctrl+c`.<br>
   This will display a "leaving" message for all other participants.

To stop the server, interrupt the terminal running the server process. If done when clients are connected, the clients will display an error. You may stop the clients like normal if that happens.
