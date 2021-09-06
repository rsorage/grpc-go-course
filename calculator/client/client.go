package main

import (
	"context"
	"fmt"
	"log"

	"github.com/rsorage/grpc-go-course/calculator/calculatorpb"
	"google.golang.org/grpc"
)

func main() {
	cc, err := grpc.Dial("0.0.0.0:50052", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}

	defer cc.Close()

	c := calculatorpb.NewCalculatorServiceClient(cc)
	doUnary(c)
}

func doUnary(c calculatorpb.CalculatorServiceClient) {
	request := &calculatorpb.SumRequest{
		A: 5,
		B: 6,
	}

	response, err := c.Sum(context.Background(), request)
	if err != nil {
		log.Fatalf("Unable to sum: %v", err)
	}

	fmt.Printf("The result is %d", response.Result)

}
