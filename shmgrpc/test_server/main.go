package main

import (
	"github.com/fullstorydev/grpchan/grpchantesting"
	"github.com/fullstorydev/grpchan/shmgrpc"
)

func main() {

	requestShmid, requestShmaddr := shmgrpc.InitializeShmRegion(shmgrpc.RequestKey, shmgrpc.Size, uintptr(shmgrpc.ServerSegFlag))
	responseShmid, responseShmaddr := shmgrpc.InitializeShmRegion(shmgrpc.ResponseKey, shmgrpc.Size, uintptr(shmgrpc.ServerSegFlag))

	qi := shmgrpc.QueueInfo{
		RequestShmid:    requestShmid,
		RequestShmaddr:  requestShmaddr,
		ResponseShmid:   responseShmid,
		ResponseShmaddr: responseShmaddr,
	}
	// svr := &grpchantesting.TestServer{}
	svc := &grpchantesting.TestServer{}
	svr := shmgrpc.NewServer(&qi, "/")

	//Register Server and instantiate with necessary information
	//Server can create queue
	//Server Can have
	go grpchantesting.RegisterTestServiceServer(svr, svc)

	//Begin handling methods from shm queue
	go svr.HandleMethods(svc)

	defer svr.Stop()

	defer shmgrpc.Detach(requestShmaddr)
	defer shmgrpc.Detach(responseShmaddr)

	defer shmgrpc.Remove(requestShmid)
	defer shmgrpc.Remove(responseShmid)

}
