package main

import (
	"chitchat/proto"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"time"

	"google.golang.org/grpc"
)

// we might need to use bidirectional streams between each client and the server
// https://grpc.io/docs/what-is-grpc/core-concepts/

// Server saves just the logical time number, to allow for joining and leaving messages

type Server struct {
	proto.UnimplementedChitChatServer
	id   int
	port int
}

var clientsConnected int
var currentTimestamp int
var clientStreams []proto.ChitChat_MessagesServer

func (server *Server) Messages(stream proto.ChitChat_MessagesServer) error {
	clientStreams = append(clientStreams, stream)

	go func() {
		for {
			msg, err := stream.Recv()
			if err == io.EOF {
				// read done.
				fmt.Println("read done")
			}
			if err != nil {
				fmt.Println(err)
				return
			}
			if err == nil {
				go DistributeMessages(msg)
			}
		}
	}()

	time.Sleep(time.Hour) // keep alive for one hour
	return nil
}

func (s *Server) Connect(ctx context.Context, in *proto.Empty) (*proto.ConnectMessage, error) {
	clientsConnected++
	UpdateTimestamp(currentTimestamp)
	return &proto.ConnectMessage{
		ClientId:  int64(clientsConnected),
		Timestamp: int64(currentTimestamp),
	}, nil
}

func DistributeMessages(message *proto.Message) {
	UpdateTimestamp(int(message.Timestamp))
	for i := 0; i < len(clientStreams); i++ {
		clientStreams[i].Send(message)
	}
}

func UpdateTimestamp(newTime int) {
	currentTimestamp = max(currentTimestamp, newTime) + 1
}

func SendWelcomeMessage(clientId int) {
	UpdateTimestamp(currentTimestamp)
	msg := &proto.Message{
		ClientId:  0,
		Timestamp: int64(currentTimestamp),
		Message:   "Client " + strconv.Itoa(clientId) + " joined",
	}
	DistributeMessages(msg)
}

func main() {
	server := &Server{
		id:   1,
		port: 5400,
	}
	clientsConnected = 0
	currentTimestamp = 0

	clientStreams = make([]proto.ChitChat_MessagesServer, 0)

	startServer(server)

	time.Sleep(time.Hour)
}

// code adapted from TAs
// https://github.com/Mai-Sigurd/grpcTimeRequestExample?tab=readme-ov-file#setting-up-the-files
func startServer(server *Server) {
	// Create a new grpc server
	grpcServer := grpc.NewServer()

	// Make the server listen at the given port (convert int port to string)
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(server.port))

	if err != nil {
		log.Fatalf("Could not create the server %v", err)
	}
	log.Printf("Started server at port: %d\n", server.port)

	// Register the grpc server and serve its listener
	proto.RegisterChitChatServer(grpcServer, server)
	serveError := grpcServer.Serve(listener)
	if serveError != nil {
		log.Fatalf("Could not serve listener")
	}
}
