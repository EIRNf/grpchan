package shmgrpc_test

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/fullstorydev/grpchan/shmgrpc"
	"github.com/fullstorydev/grpchan/test_hello_service"
)

func TestGrpcOverSharedMemory(t *testing.T) {

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
	//Server can create queue
	//Server Can have
	go test_hello_service.RegisterTestServiceServer(svr, svc)

	//Begin handling methods from shm queue
	go svr.HandleMethods(svc)

	//Placeholder URL????
	u, err := url.Parse(fmt.Sprintf("http://127.0.0.1:8080"))
	if err != nil {
		t.Fatalf("failed to parse base URL: %v", err)
	}

	// Construct Channel with necessary parameters to talk to the Server
	cc := shmgrpc.Channel{
		BaseURL:      u,
		ShmQueueInfo: &qi,
	}

	test_hello_service.RunChannelTestCases(t, &cc, true)

	svr.Stop()

	defer shmgrpc.Detach(requestShmaddr)
	defer shmgrpc.Detach(responseShmaddr)

	defer shmgrpc.Remove(requestShmid)
	defer shmgrpc.Remove(responseShmid)
}
