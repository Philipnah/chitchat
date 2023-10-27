package main

import (
	"bufio"
	"chitchat/proto"
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Clients are participants
// clients keep track of lamport time
// keep list of all messages, insertion sort when new message is recieved, to allow for message reordering

var id int
var currentTimestamp int

func main() {
	conn, _ := grpc.Dial(":5400", grpc.WithTransportCredentials(insecure.NewCredentials()))
	defer conn.Close()

	client := proto.NewChitChatClient(conn)

	Connect(client)

	stream, _ := client.Messages(context.Background())

	go Receive(stream)

	for {
		reader := bufio.NewReader(os.Stdin)
		messageText, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
		}

		msg := CreateMessage(messageText)
		stream.Send(&msg)
	}
}

func CreateMessage(message string) proto.Message {
	UpdateTimestamp(currentTimestamp)
	return proto.Message{
		ClientId:  int64(id),
		Timestamp: int64(currentTimestamp),
		Message:   message,
	}
}

func Connect(client proto.ChitChatClient) {
	connection, err := client.Connect(context.Background(), &proto.Empty{})
	if err != nil {
		fmt.Print(err)
	}
	id = int(connection.ClientId)
	currentTimestamp = int(connection.Timestamp)
}

func Receive(stream proto.ChitChat_MessagesClient) { // maybe pointer stuff?
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			fmt.Println(err)
			return
		}
		if err != nil {
			fmt.Println(err)
			return
		}
		log.Println(in.ClientId, in.Timestamp, in.Message) // maybe print all previous messages aswell

		UpdateTimestamp(int(in.Timestamp))
	}
}

func UpdateTimestamp(newTime int) {
	currentTimestamp = max(currentTimestamp, newTime) + 1
}
