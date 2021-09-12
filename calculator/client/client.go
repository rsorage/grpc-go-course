package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/rsorage/grpc-go-course/calculator/calculatorpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func main() {
	cc, err := grpc.Dial("0.0.0.0:50052", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}

	defer cc.Close()

	c := calculatorpb.NewCalculatorServiceClient(cc)

	// doUnary(c)
	// doServerStreaming(c)
	// calcAverage(c)
	doBiDiStreaming(c)
	// calcSquareRoot(c)
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

	fmt.Printf("The result is %d\n", response.Result)

}

func doServerStreaming(c calculatorpb.CalculatorServiceClient) {
	request := &calculatorpb.DecomposePrimeNumberRequest{
		Number: 120,
	}

	stream, err := c.DecomposePrimeNumber(context.Background(), request)
	if err != nil {
		log.Fatalf("Impossible to open stream: %v", err)
	}

	for {
		response, err := stream.Recv()
		if err == io.EOF {
			// Server closed stream
			break
		}
		if err != nil {
			log.Fatalf("Error while receiving stream: %v\n", err)
		}

		log.Printf("Factor received: %d\n", response.GetResult())
	}
}

func calcAverage(c calculatorpb.CalculatorServiceClient) {
	numbers := []int32{8, 56, 21, 27, 14}

	stream, err := c.Average(context.Background())
	if err != nil {
		log.Fatalf("Error opening stream to calculate average: %v\n", err)
	}

	for _, num := range numbers {
		stream.Send(&calculatorpb.AverageRequest{
			Number: num,
		})
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Error closing stream: %v\n", err)
	}

	log.Printf("Received average: %f", res.GetResult())
}

func doBiDiStreaming(c calculatorpb.CalculatorServiceClient) {

	stream, err := c.FindMaximum(context.Background())
	if err == io.EOF {
		return
	}
	if err != nil {
		return
	}

	numbers := []int32{1, 5, 3, 6, 2, 20}

	ch := make(chan struct{})

	// Send stream
	go func() {
		for _, n := range numbers {
			log.Printf("Sending number: %v\n", n)
			err = stream.Send(&calculatorpb.FindMaximumRequest{
				Number: n,
			})

			time.Sleep(200 * time.Millisecond)

			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Error sending msg to stream: %v", err)
				break
			}
		}

		stream.CloseSend()
	}()

	// Receive stream
	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				break
			}

			log.Printf("New max received: %v\n", res.GetMax())
		}
		close(ch)
	}()

	<-ch
}

func calcSquareRoot(c calculatorpb.CalculatorServiceClient) {
	number := -8.0

	request := &calculatorpb.SquareRootRequest{
		Number: number,
	}

	response, err := c.SquareRoot(context.Background(), request)
	if err != nil {
		respErr, ok := status.FromError(err)
		if ok {
			// actual error from gRPC
			log.Printf("Server sent --> %s", respErr.Message())
			if respErr.Code() == codes.InvalidArgument {
				log.Println("We probably sent a negative number!")
			}
			return
		} else {
			log.Fatalf("Error calculating square root: %v", err)
			return
		}
	}

	log.Printf("Square root of %f is %f\n", number, response.GetResult())
}
