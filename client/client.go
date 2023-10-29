package main

import (
	"bufio"
	"chitchat/proto"
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Clients are participants
// clients keep track of lamport time
// keep list of all messages, insertion sort when new message is recieved, to allow for message reordering

var id int
var currentTimestamp int
var messagesReceived []proto.Message

func main() {
	messagesReceived = make([]proto.Message, 0)
	conn, _ := grpc.Dial(":5400", grpc.WithTransportCredentials(insecure.NewCredentials()))
	defer conn.Close()

	logFile, _ := os.Create("logfile" + strconv.Itoa(rand.Int()) + ".log")
	defer logFile.Close()
	log.SetOutput(logFile)

	client := proto.NewChitChatClient(conn)

	Connect(client)

	stream, _ := client.Messages(context.Background())

	go Receive(stream)

	for {
		reader := bufio.NewReader(os.Stdin)
		rawMessageText, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
		}

		messageText := strings.Trim(rawMessageText, "\n")

		messageError := CheckMessageRequirements(messageText)
		if messageError == "toolong" {
			fmt.Println("Your message was too long")
		} else if messageError == "tooshort" {
			fmt.Println("Your message was too short")
		} else {
			msg := CreateMessage(messageText)
			stream.Send(&msg)
		}
	}
}

func CheckMessageRequirements(msgText string) string {
	if len(msgText) > 128 {
		return "toolong"
	}
	if len(msgText) <= 0 {
		return "tooshort"
	}
	return ""
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

func Receive(stream proto.ChitChat_MessagesClient) {
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
		log.Println(FormatMessage(in))
		messagesReceived = append(messagesReceived, *in)
		UpdateTimestamp(int(in.Timestamp))
		PrintMessages()
	}
}

func PrintMessages() {
	ClearTerminal()
	sort.SliceStable(messagesReceived, func(i, j int) bool {
		if messagesReceived[i].Timestamp != messagesReceived[j].Timestamp {
			return messagesReceived[i].Timestamp < messagesReceived[j].Timestamp
		} else {
			return messagesReceived[i].ClientId < messagesReceived[j].ClientId
		}
	})

	for i := 0; i < len(messagesReceived); i++ {
		message := &messagesReceived[i]
		fmt.Println(FormatMessage(message))
	}
}

func FormatMessage(message *proto.Message) string {
	return ("Event(" + strconv.Itoa(int(message.Timestamp)) + ", " + strconv.Itoa(int(message.ClientId)) + ") " + message.Message)
}

func ClearTerminal() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func UpdateTimestamp(newTime int) {
	currentTimestamp = max(currentTimestamp, newTime) + 1
}
