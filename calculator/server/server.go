package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"net"

	"github.com/rsorage/grpc-go-course/calculator/calculatorpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct{}

func (*server) Sum(ctx context.Context, req *calculatorpb.SumRequest) (*calculatorpb.SumResponse, error) {
	result := int64(req.A) + int64(req.B)

	response := &calculatorpb.SumResponse{
		Result: result,
	}

	return response, nil
}

func (*server) DecomposePrimeNumber(req *calculatorpb.DecomposePrimeNumberRequest, stream calculatorpb.CalculatorService_DecomposePrimeNumberServer) error {
	k := int32(2)
	number := req.GetNumber()

	for number > 1 {
		if number%k == 0 {
			response := &calculatorpb.DecomposePrimeNumberResponse{
				Result: k,
			}

			stream.Send(response)
			number /= k
		} else {
			k += 1
		}
	}

	return nil
}

func (*server) Average(stream calculatorpb.CalculatorService_AverageServer) error {

	numbers := []int32{}

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			log.Printf("Calculating average of: %v", numbers)
			avg := calcAverage(numbers)
			return stream.SendAndClose(&calculatorpb.AverageResponse{
				Result: avg,
			})
		}
		if err != nil {
			log.Printf("Error receiving stream: %v", err)
		}
		log.Printf("Request received: %v", req)
		numbers = append(numbers, req.GetNumber())
	}
}

func calcAverage(numbers []int32) float64 {
	var sum int32 = 0

	for _, num := range numbers {
		sum += num
	}

	return float64(sum) / float64(len(numbers))
}

func (*server) SquareRoot(ctx context.Context, req *calculatorpb.SquareRootRequest) (*calculatorpb.SquareRootResponse, error) {
	log.Printf("Receiving call to SquareRoot with: %v\n", req)
	number := req.GetNumber()

	if number < 0 {
		err := status.Errorf(codes.InvalidArgument, fmt.Sprintf("Received a negative number: %v", number))
		return nil, err
	}

	response := &calculatorpb.SquareRootResponse{
		Result: math.Sqrt(number),
	}

	return response, nil
}

func main() {
	lis, err := net.Listen("tcp", "0.0.0.0:50052")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	calculatorpb.RegisterCalculatorServiceServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
