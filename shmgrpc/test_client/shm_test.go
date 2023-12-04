package shmgrpc_test

import (
	"testing"

	"github.com/fullstorydev/grpchan/shmgrpc"
	"github.com/fullstorydev/grpchan/test_hello_service"
)

func TestGrpcOverSharedMemory(t *testing.T) {
	cc := shmgrpc.NewChannel("localhost", "http://127.0.0.1:8080/hello")

	test_hello_service.RunChannelTestCases(t, cc, true)
}
