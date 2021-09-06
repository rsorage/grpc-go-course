package main

import (
	"context"
	"fmt"
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

	fmt.Println("Response from Greet: ", ret.Result)
}
