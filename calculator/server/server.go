package main

import (
	"context"
	"log"
	"net"

	"github.com/rsorage/grpc-go-course/calculator/calculatorpb"
	"google.golang.org/grpc"
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
