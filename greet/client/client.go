package main

import (
	"context"
	"io"
	"log"

	"github.com/rsorage/grpc-go-course/greet/greetpb"
	"google.golang.org/grpc"
)

func main() {

	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	defer cc.Close()

	c := greetpb.NewGreetServiceClient(cc)

	doUnary(c)
	doServerStreaming(c)
}

func doUnary(c greetpb.GreetServiceClient) {

	request := &greetpb.GreetRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Ramon",
			LastName:  "Sorage",
		},
	}

	ret, err := c.Greet(context.Background(), request)
	if err != nil {
		log.Fatalf("Unable to greet: %v", err)
	}
	log.Println("Response from Greet: ", ret.Result)
}

func doServerStreaming(c greetpb.GreetServiceClient) {

	request := &greetpb.GreetManyTimesRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Ramon",
			LastName:  "Sorage",
		},
	}

	stream, err := c.GreetManyTimes(context.Background(), request)
	if err != nil {
		log.Fatalf("Unable to greet many times: %v", err)
	}

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			// server closed stream
			break
		}
		if err != nil {
			log.Fatalf("Error while reading stream: %v", err)
		}
		log.Printf("Response from GreetManyTimes: %v", msg.GetResult())
	}
}
