package main

import (
	"context"
	"io"
	"log"
	"time"

	"github.com/rsorage/grpc-go-course/greet/greetpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func main() {

	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	defer cc.Close()

	c := greetpb.NewGreetServiceClient(cc)

	// doUnary(c)
	// doServerStreaming(c)
	doUnaryWithDeadline(c, 5*time.Second)
	doUnaryWithDeadline(c, 1*time.Second)
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

func doUnaryWithDeadline(c greetpb.GreetServiceClient, timeout time.Duration) {
	log.Println("Starting to do unary with deadline RPC...")
	req := &greetpb.GreetWithDeadlineRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Ramon",
			LastName:  "Sorage",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	res, err := c.GreetWithDeadline(ctx, req)
	if err != nil {
		statusErr, ok := status.FromError(err)

		if ok {
			if statusErr.Code() == codes.DeadlineExceeded {
				log.Println("Timeout was hit! Deadline exceeded...")
			} else {
				log.Printf("Unexpected error: %v\n", statusErr)
			}
		} else {
			log.Fatalf("Error while calling GreetWithDeadline RPC: %v", err)
		}
		return
	}

	log.Printf("Response from GreetWithDeadline: %v", res.Result)
}
