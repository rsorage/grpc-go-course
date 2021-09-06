package main

import (
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

	c := greetpb.NewGreetServiceClient((cc))
	fmt.Printf("Created: %f", c)
}
