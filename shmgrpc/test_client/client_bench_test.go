package shmgrpc_test

import (
	"net/url"
	"testing"

	"github.com/fullstorydev/grpchan/grpchantesting"
	"github.com/fullstorydev/grpchan/shmgrpc"
)

func BenchmarkGrpcOverSharedMemory(b *testing.B) {

	u, err := url.Parse("http://127.0.0.1:8080")
	if err != nil {
		b.Fatalf("failed to parse base URL: %v", err)
	}

	// Construct Channel with necessary parameters to talk to the Server
	cc := shmgrpc.NewChannel(u, "hello")

	// grpchantesting.RunChannelTestCases(t, &cc, true)
	grpchantesting.RunChannelBenchmarkCases(b, cc, false)

	defer shmgrpc.StopPollingQueue(shmgrpc.GetQueue(cc.ShmQueueInfo.RequestShmaddr))
	defer shmgrpc.StopPollingQueue(shmgrpc.GetQueue(cc.ShmQueueInfo.ResponseShmaddr))

	defer shmgrpc.Detach(cc.ShmQueueInfo.RequestShmaddr)
	defer shmgrpc.Detach(cc.ShmQueueInfo.ResponseShmaddr)

}
