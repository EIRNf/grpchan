package test_hello_service

//go:generate protoc --proto_path=../ --go_out=./ --go-grpc_out=./ --grpchan_out=legacy_stubs:./ grpchantesting/test.proto

import (
	"context"
)

// TestServer has default responses to the various kinds of methods.
type TestServer struct {
	UnimplementedTestServiceServer
}

func (s *TestServer) SayHello(ctx context.Context, in *HelloRequest) (*HelloReply, error) {
	// log.Printf("Received: %v", in.GetName())
	return &HelloReply{Message: "Hello " + in.GetName()}, nil
}

func (s *TestServer) SayGoodbye(ctx context.Context, in *GoodbyeRequest) (*GoodbyeReply, error) {
	// log.Printf("Received: %v", in.GetName())
	return &GoodbyeReply{Message: "Goodbye " + in.GetName()}, nil
}
