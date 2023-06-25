package shmgrpc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"syscall"

	"github.com/fullstorydev/grpchan"
	"github.com/fullstorydev/grpchan/internal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/encoding"
	grpcproto "google.golang.org/grpc/encoding/proto"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/siadat/ipc"
)

// Server is grpc over shared memory. It
// acts as a grpc ServiceRegistrar
type Server struct {
	handlers         grpchan.HandlerMap
	basePath         string
	ShmQueueInfo     *QueueInfo
	opts             handlerOpts
	unaryInterceptor grpc.UnaryServerInterceptor
}

var _ grpc.ServiceRegistrar = (*Server)(nil)

// ServerOption is an option used when constructing a NewServer.
type ServerOption interface {
	apply(*Server)
}

type serverOptFunc func(*Server)

func (fn serverOptFunc) apply(s *Server) {
	fn(s)
}

type handlerOpts struct {
	errFunc func(context.Context, *status.Status, http.ResponseWriter)
}

func NewServer(ShmQueueInfo *QueueInfo, basePath string, opts ...ServerOption) *Server {
	var s Server
	s.basePath = basePath
	s.handlers = grpchan.HandlerMap{}
	s.ShmQueueInfo = ShmQueueInfo

	var key, qid uint
	var err error

	//Instantiate queue for processing
	key, err = ipc.Ftok(s.ShmQueueInfo.QueuePath, s.ShmQueueInfo.QueueId)
	if err != nil {
		panic(fmt.Sprintf("SERVER: Failed to generate key: %s\n", err))
	} else {
		// fmt.Printf("SERVER: Generate key %d\n", key)
	}

	qid, err = ipc.Msgget(key, ipc.IPC_CREAT|0666)
	if err != nil {
		panic(fmt.Sprintf("SERVER: Failed to create ipc key %d: %s\n", key, err))
	} else {
		// fmt.Printf("SERVER: Create ipc queue id %d\n", qid)
	}

	s.ShmQueueInfo.Qid = qid

	return &s
}

// Attempt of a HandlerFunc
// type HandlerFunc func(Response, *http.Request)

// func (f HandlerFunc) ServeSHM() {
// 	f()
// }

func (s *Server) RegisterService(desc *grpc.ServiceDesc, svr interface{}) {

	s.handlers.RegisterService(desc, svr)
	s.unaryInterceptor = nil

	qid := s.ShmQueueInfo.Qid

	for {
		//Check for valid meta data message
		msg_req := &ipc.Msgbuf{
			Mtype: s.ShmQueueInfo.QueueReqTypeMeta}
		//Mesrcv on response message type
	retry:
		err := ipc.Msgrcv(qid, msg_req, 0)
		if err != nil || msg_req.Mtext == nil {
			if err == syscall.EINTR {
				//Try again????
				goto retry
			}
			panic(fmt.Sprintf("SERVER: Failed to receive message to ipc id %d: %s\n", qid, err))
		} else {
			// fmt.Printf("SERVER: Message %v receive to ipc id %d\n", msg_req.Mtext, qid)
		}

		//Parse bytes into object
		var message_req_meta ShmMessage
		json.Unmarshal(msg_req.Mtext, &message_req_meta)

		payload_buffer := unsafeGetBytes(message_req_meta.Payload)

		fullName := message_req_meta.Method
		strs := strings.SplitN(fullName[1:], "/", 2)
		serviceName := strs[0]
		methodName := strs[1]
		clientCtx := message_req_meta.Context
		if clientCtx == nil { // Temp in case of empty context.
			clientCtx = context.Background()
		}

		clientCtx, cancel, err := contextFromHeaders(clientCtx, message_req_meta.Headers)
		if err != nil {
			// writeError(w, http.StatusBadRequest)
			return
		}

		defer cancel()

		//Get Service Descriptor and Handler
		sd, handler := s.handlers.QueryService(serviceName)
		if sd == nil {
			// service name not found
			status.Errorf(codes.Unimplemented, "service %s not implemented", message_req_meta.Method)
		}

		//Get Method Descriptor
		md := internal.FindUnaryMethod(methodName, sd.Methods)
		if md == nil {
			// method name not found
			status.Errorf(codes.Unimplemented, "method %s/%s not implemented", serviceName, methodName)
		}

		//Get Codec for content type.
		codec := encoding.GetCodec(grpcproto.Name)

		// Function to unmarshal payload using proto
		dec := func(msg interface{}) error {
			val := payload_buffer
			if err := codec.Unmarshal(val, msg); err != nil {
				return status.Error(codes.InvalidArgument, err.Error())
			}
			return nil
		}

		// Implements server transport stream
		sts := internal.UnaryServerTransportStream{Name: methodName}
		ctx := grpc.NewContextWithServerTransportStream(makeServerContext(clientCtx), &sts)

		//Get resp write back
		resp, err := md.Handler(
			handler,
			ctx,
			dec,
			s.unaryInterceptor)
		if err != nil {
			status.Errorf(codes.Unknown, "Handler error: %s ", err.Error())
			//TODO: Error code must be sent back to client
		}

		resp_buffer, err := codec.Marshal(resp)
		if err != nil {
			status.Errorf(codes.Unknown, "Codec Marshalling error: %s ", err.Error())
		}

		message_resp := &ShmMessage{
			Method:   methodName,
			Context:  ctx,
			Headers:  sts.GetHeaders(),
			Trailers: sts.GetTrailers(),
			Payload:  ByteSlice2String(resp_buffer),
		}

		serialized_resp, err := json.Marshal(message_resp)
		if err != nil {
			status.Errorf(codes.Unknown, "Codec Marshalling error: %s ", err.Error())
		}

		//Construct respond buffer (this will have to change)
		msg_resp := &ipc.Msgbuf{
			Mtype: s.ShmQueueInfo.QueueRespTypeMeta,
			Mtext: serialized_resp,
		}
		//Write back
		err = ipc.Msgsnd(qid, msg_resp, 0)
		if err != nil {
			panic(fmt.Sprintf("SERVER: Failed to send resp to ipc id %d: %s\n", qid, err))
		} else {
			// fmt.Printf("SERVER:Message %v send to ipc id %d\n", msg_req, qid)
		}
	}

}

// contextFromHeaders returns a child of the given context that is populated
// using the given headers. The headers are converted to incoming metadata that
// can be retrieved via metadata.FromIncomingContext. If the headers contain a
// GRPC timeout, that is used to create a timeout for the returned context.
func contextFromHeaders(parent context.Context, md metadata.MD) (context.Context, context.CancelFunc, error) {
	cancel := func() {} // default to no-op

	ctx := metadata.NewIncomingContext(parent, md)

	// deadline propagation
	// timeout := h.Get("GRPC-Timeout")
	// if timeout != "" {
	// 	// See GRPC wire format, "Timeout" component of request: https://grpc.io/docs/guides/wire.html#requests
	// 	suffix := timeout[len(timeout)-1]
	// 	if timeoutVal, err := strconv.ParseInt(timeout[:len(timeout)-1], 10, 64); err == nil {
	// 		var unit time.Duration
	// 		switch suffix {
	// 		case 'H':
	// 			unit = time.Hour
	// 		case 'M':
	// 			unit = time.Minute
	// 		case 'S':
	// 			unit = time.Second
	// 		case 'm':
	// 			unit = time.Millisecond
	// 		case 'u':
	// 			unit = time.Microsecond
	// 		case 'n':
	// 			unit = time.Nanosecond
	// 		}
	// 		if unit != 0 {
	// 			ctx, cancel = context.WithTimeout(ctx, time.Duration(timeoutVal)*unit)
	// 		}
	// 	}
	// }
	return ctx, cancel, nil
}

// noValuesContext wraps a context but prevents access to its values. This is
// useful when you need a child context only to propagate cancellations and
// deadlines, but explicitly *not* to propagate values.
type noValuesContext struct {
	context.Context
}

func makeServerContext(ctx context.Context) context.Context {
	// We don't want the server have any of the values in the client's context
	// since that can inadvertently leak state from the client to the server.
	// But we do want a child context, just so that request deadlines and client
	// cancellations work seamlessly.
	newCtx := context.Context(noValuesContext{ctx})

	if meta, ok := metadata.FromOutgoingContext(ctx); ok {
		newCtx = metadata.NewIncomingContext(newCtx, meta)
	}
	// newCtx = peer.NewContext(newCtx, &inprocessPeer)
	// newCtx = context.WithValue(newCtx, &clientContextKey, ctx)
	return newCtx
}

// func handleMethod(svr interface{}, serviceName string, desc *grpc.MethodDesc) func() {

// 	fullMethod := fmt.Sprintf("/%s/%s", serviceName, desc.MethodName)
// 	fmt.Println(fullMethod)
// 	return f()
// }