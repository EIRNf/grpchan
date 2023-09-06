package test_hello_service

import (
	"bytes"
	"context"
	"flag"
	"net"
	"testing"
	"time"

	"github.com/fullstorydev/grpchan/grpchantesting"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// We test all of our channel test cases by running them against a normal
// *grpc.Server and *grpc.ClientConn, to make sure they are asserting the
// same behavior exhibited by the standard HTTP/2 channel implementation.
func TestHelloChannelTestCases(t *testing.T) {

	s := grpc.NewServer()

	RegisterTestServiceServer(s, &TestServer{})

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen on socket: %v", err)
	}
	go s.Serve(l)
	defer s.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addr := l.Addr().String()
	cc, err := grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithBlock(), grpc.FailOnNonTempDialError(true))
	if err != nil {
		t.Fatalf("failed to dial address: %s", addr)
	}
	defer cc.Close()

	// RunChannelTestCases(t, cc, true)

	var defaultName = "world"
	var testPayload = []byte{100, 90, 80, 70, 60, 50, 40, 30, 20, 10, 0}
	var testOutgoingMd = map[string][]byte{
		"foo":        []byte("bar"),
		"baz":        []byte("bedazzle"),
		"pickle-bin": testPayload,
	}

	cli := NewTestServiceClient(cc)
	cli_ctx := metadata.NewOutgoingContext(context.Background(), grpchantesting.MetadataNew(testOutgoingMd))

	name := flag.String("name", defaultName, "Name to greet")

	t.Run("success", func(t *testing.T) {
		req := &HelloRequest{Name: *name}
		rsp, err := cli.SayHello(cli_ctx, req)
		if err != nil {
			t.Fatalf("RPC failed: %v", err)
		}
		if !bytes.Equal([]byte("Hello world"), []byte(rsp.GetMessage())) {
			t.Fatalf("wrong payload returned: expecting %v; got %v", testPayload, rsp.GetMessage())
		}

	})
}
