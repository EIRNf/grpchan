package grpchantesting

//go:generate protoc --proto_path=../ --go_out=./ --go-grpc_out=./ --grpchan_out=legacy_stubs:./ grpchantesting/test.proto

import (
	"context"
	"io"
	"time"

	spb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	empty "google.golang.org/protobuf/types/known/emptypb"
)

// TestServer has default responses to the various kinds of methods.
type TestServer struct {
	UnimplementedTestServiceServer
}

// Unary implements the TestService server interface.
func (s *TestServer) Unary(ctx context.Context, req *Message) (*Message, error) {
	if req.DelayMillis > 0 {
		time.Sleep(time.Millisecond * time.Duration(req.DelayMillis))
	}
	grpc.SetHeader(ctx, MetadataNew(req.Headers))
	grpc.SetTrailer(ctx, MetadataNew(req.Trailers))
	if req.Code != 0 {
		return nil, statusFromRequest(req)
	}
	md, _ := metadata.FromIncomingContext(ctx)
	return &Message{
		Headers: asMap(md),
		Payload: req.Payload,
	}, nil
}

func statusFromRequest(req *Message) error {
	statProto := spb.Status{
		Code:    req.Code,
		Message: "error",
		Details: req.ErrorDetails,
	}
	return status.FromProto(&statProto).Err()
}

// ClientStream implements the TestService server interface.
func (s *TestServer) ClientStream(cs TestService_ClientStreamServer) error {
	var req *Message
	count := int32(0)
	for {
		r, err := cs.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		req = r
		count++
		if req.Code != 0 {
			break
		}
	}
	if req == nil {
		req = &Message{}
	}
	if req.DelayMillis > 0 {
		time.Sleep(time.Millisecond * time.Duration(req.DelayMillis))
	}
	if err := cs.SetHeader(MetadataNew(req.Headers)); err != nil {
		return err
	}
	cs.SetTrailer(MetadataNew(req.Trailers))
	if req.Code != 0 {
		return statusFromRequest(req)
	}
	md, _ := metadata.FromIncomingContext(cs.Context())
	return cs.SendAndClose(&Message{
		Headers: asMap(md),
		Payload: req.Payload,
		Count:   count,
	})
}

// ServerStream implements the TestService server interface.
func (s *TestServer) ServerStream(req *Message, ss TestService_ServerStreamServer) error {
	if req.DelayMillis > 0 {
		time.Sleep(time.Millisecond * time.Duration(req.DelayMillis))
	}
	md, _ := metadata.FromIncomingContext(ss.Context())
	if err := ss.SetHeader(MetadataNew(req.Headers)); err != nil {
		return err
	}
	for i := 0; i < int(req.Count); i++ {
		err := ss.Send(&Message{
			Headers: asMap(md),
			Payload: req.Payload,
		})
		if err != nil {
			return err
		}
	}
	ss.SetTrailer(MetadataNew(req.Trailers))
	if req.Code != 0 {
		return statusFromRequest(req)
	}
	return nil
}

// BidiStream implements the TestService server interface.
func (s *TestServer) BidiStream(str TestService_BidiStreamServer) error {
	md, _ := metadata.FromIncomingContext(str.Context())
	var req *Message
	count := int32(0)
	var responses []*Message
	isHalfDuplex := false
	for {
		r, err := str.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		req = r
		if req.DelayMillis > 0 {
			time.Sleep(time.Millisecond * time.Duration(req.DelayMillis))
		}
		if count == 0 {
			if err := str.SetHeader(MetadataNew(req.Headers)); err != nil {
				return err
			}
			isHalfDuplex = req.Count < 0
		}
		count++
		if req.Code != 0 {
			break
		}
		replyMsg := &Message{
			Headers: asMap(md),
			Payload: req.Payload,
			Count:   count,
		}
		if isHalfDuplex {
			// half duplex means we fully consume the client stream before we
			// start sending responses, so buffer these messages in a slice
			responses = append(responses, replyMsg)
		} else if err = str.Send(replyMsg); err != nil {
			return err
		}
	}
	if isHalfDuplex {
		// now we can send out all buffered responses
		for _, response := range responses {
			if err := str.Send(response); err != nil {
				return err
			}
		}
	}
	if req != nil {
		str.SetTrailer(MetadataNew(req.Trailers))
		if req.Code != 0 {
			return statusFromRequest(req)
		}
	}
	return nil
}

// UseExternalMessageTwice implements the TestService server interface.
func (s *TestServer) UseExternalMessageTwice(ctx context.Context, in *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}

func asMap(md metadata.MD) map[string][]byte {
	m := map[string][]byte{}
	for k, vs := range md {
		if len(vs) == 0 {
			continue
		}
		m[k] = []byte(vs[len(vs)-1])
	}
	return m
}
