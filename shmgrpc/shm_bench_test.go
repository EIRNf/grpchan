package shmgrpc_test

import (
	"net/url"
	"testing"

	"github.com/fullstorydev/grpchan/shmgrpc"
	"github.com/fullstorydev/grpchan/test_hello_service"
)

func BenchmarkGrpcOverSharedMemory(b *testing.B) {

	requestShmid, requestShmaddr := shmgrpc.InitializeShmRegion(shmgrpc.RequestKey, shmgrpc.Size, uintptr(shmgrpc.ServerSegFlag))
	responseShmid, responseShmaddr := shmgrpc.InitializeShmRegion(shmgrpc.ResponseKey, shmgrpc.Size, uintptr(shmgrpc.ServerSegFlag))

	qi := shmgrpc.QueueInfo{
		RequestShmid:    requestShmid,
		RequestShmaddr:  requestShmaddr,
		ResponseShmid:   responseShmid,
		ResponseShmaddr: responseShmaddr,
	}

	// svr := &grpchantesting.TestServer{}
	svc := &test_hello_service.TestServer{}
	svr := shmgrpc.NewServer(&qi, "/")

	//Register Server and instantiate with necessary information
	go test_hello_service.RegisterTestServiceServer(svr, svc)

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
		ShmQueueInfo: &qi,
	}

	// grpchantesting.RunChannelTestCases(t, &cc, true)
	test_hello_service.RunChannelBenchmarkCases(b, &cc, false)

	svr.Stop()

	defer shmgrpc.Detach(requestShmaddr)
	defer shmgrpc.Detach(responseShmaddr)

	defer shmgrpc.Remove(requestShmid)
	defer shmgrpc.Remove(responseShmid)
}
