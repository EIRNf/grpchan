package shmgrpc_test

import (
	"net/url"
	"testing"

	"github.com/fullstorydev/grpchan/shmgrpc"
	"github.com/fullstorydev/grpchan/test_hello_service"
)

func BenchmarkGrpcOverSharedMemory(b *testing.B) {

	u, err := url.Parse("http://127.0.0.1:8080")
	if err != nil {
		b.Fatalf("failed to parse base URL: %v", err)
	}

	// Construct Channel with necessary parameters to talk to the Server
	cc := shmgrpc.NewChannel(u, "/hello")

	// grpchantesting.RunChannelTestCases(t, &cc, true)
	test_hello_service.RunChannelBenchmarkCases(b, cc, false)

	shmgrpc.StopPollingQueue(shmgrpc.GetQueue(cc.ShmQueueInfo.RequestShmaddr))
	shmgrpc.StopPollingQueue(shmgrpc.GetQueue(cc.ShmQueueInfo.ResponseShmaddr))

	defer shmgrpc.Detach(cc.ShmQueueInfo.RequestShmaddr)
	defer shmgrpc.Detach(cc.ShmQueueInfo.ResponseShmaddr)

}
