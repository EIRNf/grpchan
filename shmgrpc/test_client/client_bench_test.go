package shmgrpc_test

import (
	"os"
	"testing"

	"github.com/fullstorydev/grpchan/shmgrpc"
	"github.com/fullstorydev/grpchan/test_hello_service"

	"runtime/pprof"
)

func BenchmarkGrpcOverSharedMemory(b *testing.B) {
	f, _ := os.Create("bench.prof")

	pprof.StartCPUProfile(f)

	cc := shmgrpc.NewChannel("localhost", "http://127.0.0.1:8080/hello")

	test_hello_service.RunChannelBenchmarkCases(b, cc, false)

	defer pprof.StopCPUProfile()

}
