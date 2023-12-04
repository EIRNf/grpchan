package shmgrpc_test

import (
	"testing"

	"github.com/fullstorydev/grpchan/shmgrpc"
	"github.com/fullstorydev/grpchan/test_hello_service"
)

func BenchmarkGrpcOverSharedMemory(b *testing.B) {

	// svr := &grpchantesting.TestServer{}
	svc := &test_hello_service.TestServer{}
	svr := shmgrpc.NewServer("/hello")

	//Register Server and instantiate with necessary information
	test_hello_service.RegisterTestServiceServer(svr, svc)

	//Create Listener
	lis := shmgrpc.Listen("http://127.0.0.1:8080/hello")

	go svr.Serve(lis)
	defer svr.Stop()

	cc := shmgrpc.NewChannel("localhost", "http://127.0.0.1:8080/hello")

	// grpchantesting.RunChannelTestCases(t, &cc, true)
	test_hello_service.RunChannelBenchmarkCases(b, cc, false)

	defer svr.Stop()
}
