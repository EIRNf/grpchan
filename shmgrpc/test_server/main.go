package main

import (
	"github.com/fullstorydev/grpchan/shmgrpc"
	"github.com/fullstorydev/grpchan/test_hello_service"
)

func main() {

	// svr := &grpchantesting.TestServer{}
	svc := &test_hello_service.TestServer{}
	svr := shmgrpc.NewServer("/hello")

	//Register Server and instantiate with necessary information
	//Server can create queue
	//Server Can have
	go test_hello_service.RegisterTestServiceServer(svr, svc)

	//Begin handling methods from shm queue
	svr.HandleMethods(svc)

	defer svr.Stop()

}
