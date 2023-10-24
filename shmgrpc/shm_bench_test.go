package shmgrpc_test

import (
	"net/url"
	"sync"
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

	//Begin handling methods from shm queue
	go svr.HandleMethods(svc)

	//Placeholder URL????
	u, err := url.Parse("http://127.0.0.1:8080")
	if err != nil {
		b.Fatalf("failed to parse base URL: %v", err)
	}

	// Construct Channel with necessary parameters to talk to the Server
	cc := shmgrpc.Channel{
		BaseURL:      u,
		ShmQueueInfo: svr.ShmQueueInfo,
		Lock:         sync.Mutex{},
	}

	// grpchantesting.RunChannelTestCases(t, &cc, true)
	test_hello_service.RunChannelBenchmarkCases(b, &cc, false)

	defer svr.Stop()
}
