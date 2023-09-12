package shmgrpc_test

import (
	"net/url"
	"testing"

	"github.com/fullstorydev/grpchan/grpchantesting"
	"github.com/fullstorydev/grpchan/shmgrpc"
)

func BenchmarkGrpcOverSharedMemory(b *testing.B) {

	// svr := &grpchantesting.TestServer{}
	svc := &grpchantesting.TestServer{}
	svr := shmgrpc.NewServer("/hello")

	//Register Server and instantiate with necessary information
	grpchantesting.RegisterTestServiceServer(svr, svc)

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
	}

	// grpchantesting.RunChannelTestCases(t, &cc, true)
	grpchantesting.RunChannelBenchmarkCases(b, &cc, false)

	defer svr.Stop()
}
