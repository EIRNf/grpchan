package shmgrpc_test

import (
	"testing"

	"github.com/fullstorydev/grpchan/shmgrpc"
	"github.com/fullstorydev/grpchan/test_hello_service"
)

func BenchmarkGrpcOverSharedMemory(b *testing.B) {

	cc := shmgrpc.NewChannel("localhost", "http://127.0.0.1:8080/hello")

	test_hello_service.RunChannelBenchmarkCases(b, cc, false)

}
