package test_hello_service

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"testing"

	"github.com/aclements/go-moremath/stats"
	"github.com/loov/hrtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// RunChannelBenchmarkCases runs numerous test cases to exercise the behavior of the
// given channel. The server side of the channel needs to have a *TestServer (in
// this package) registered to provide the implementation of fsgrpc.TestService
// (proto in this package). If the channel does not support full-duplex
// communication, it must provide at least half-duplex support for bidirectional
// streams.
//
// The test cases will be defined as child tests by invoking t.Run on the given
// *testing.T.

func RunChannelBenchmarkCases(b *testing.B, ch grpc.ClientConnInterface, supportsFullDuplex bool) {
	cli := NewTestServiceClient(ch)

	b.Run("hello", func(b *testing.B) { BenchmarkHelloHistogram(b, cli) })
	// b.RunParallel(func(pb *testing.PB) { BenchmarkUnaryLatencyParallel(pb, cli) })

}

func BenchmarkUnaryLatencyParallel(pb *testing.PB, cli TestServiceClient) {
	ctx := metadata.NewOutgoingContext(context.Background(), MetadataNew(testOutgoingMd))

	var name = defaultName

	for pb.Next() {
		req := &HelloRequest{Name: name}
		rsp, err := cli.SayHello(ctx, req)

		if err != nil {
		}
		if rsp != nil {

		}

	}

}

func BenchmarkHelloHistogram(b *testing.B, cli TestServiceClient) {

	bench := hrtime.NewBenchmark(200000)

	ctx := metadata.NewOutgoingContext(context.Background(), MetadataNew(testOutgoingMd))

	name := flag.String("benchname", defaultName, "Name to greet")

	// b.Run("success", func(b *testing.B) {
	for bench.Next() {
		req := &HelloRequest{Name: *name}
		rsp, err := cli.SayHello(ctx, req)
		if err != nil {
			b.Fatalf("RPC failed: %v", err)
		}
		if !bytes.Equal([]byte("Hello world"), []byte(rsp.GetMessage())) {
			b.Fatalf("wrong payload returned: expecting %v; got %v", testPayload, rsp.GetMessage())
		}

	}

	fmt.Println(bench.Histogram(10))

	runs := bench.Float64s()

	fmt.Printf("Mean: %f\n", stats.Mean(runs)*0.001)
	fmt.Printf("StdDev: %f\n", stats.StdDev(runs)*0.001)
	fmt.Printf("NumElements: %d\n", len(runs))

}
