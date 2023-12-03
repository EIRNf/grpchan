package shmgrpc

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/fullstorydev/grpchan"
	"github.com/fullstorydev/grpchan/internal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/encoding"
	grpcproto "google.golang.org/grpc/encoding/proto"
	"google.golang.org/grpc/internal/grpcsync"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// type ServeMux struct {
// 	mu    sync.RWMutex
// 	m     map[string]muxEntry
// 	es    []muxEntry
// 	hosts bool
// }

// type muxEntry struct {
// 	h       Handler
// 	pattern string
// }

// type Handler interface {
// 	ServeMethod(,)
// }

// Server is grpc over shared memory. It
// acts as a grpc ServiceRegistrar
type Server struct {
	handlers         grpchan.HandlerMap
	basePath         string
	opts             handlerOpts
	unaryInterceptor grpc.UnaryServerInterceptor

	quit    *grpcsync.Event
	done    *grpcsync.Event
	serveWG sync.WaitGroup

	ErrorLog *log.Logger

	mu sync.Mutex

	// Listener accepting connections on a particular IP  and port
	lis *Listener

	// Map of queue pairs for boolean of active or inactive connections
	conns map[int]*QueuePair

	ShmQueueInfo  *QueueInfo
	responseQueue *Queue
	requestQeuue  *Queue
}

// "Listens" on the queue accept method and hands it off to a dedicate thread to manage.
// Handles the notnets layer
type Listener struct {
	mu              sync.Mutex
	notnets_context *ServerContext
	addr            string
}

// Instantiates notnets listening loop. From listener accept can be called
func Listen(addr string) *Listener {
	var lis Listener
	lis.notnets_context = RegisterServer(addr)
	lis.addr = addr
	return &lis
}

func (l *Listener) Accept() (*QueuePair, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	queue := l.notnets_context.Accept()
	var err error
	return queue, err
}

func (l *Listener) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.notnets_context.Shutdown()
	return nil
}

func (l *Listener) Addr() string {
	return l.Addr()
}

var _ grpc.ServiceRegistrar = (*Server)(nil)

var (
	sSerReqStruct   ShmMessage
	sSerReqWritten  bool = false
	sSerRespData    [600]byte
	sSerRespLen     int
	sSerRespWritten bool = false

	sserPayload        []byte
	sserPayloadWritten bool = false
)

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

func NewServer(basePath string, opts ...ServerOption) *Server {
	var s Server
	s.basePath = basePath
	s.handlers = grpchan.HandlerMap{}
	for _, o := range opts {
		o.apply(&s)
	}

	s.conns = make(map[int]*QueuePair)

	return &s
}

// Accepts incoming connections on the listener lis,
// creating new shm connections and a dedicated
// goroutine fo reach and then call the registered
// handles for then.
func (s *Server) Serve(lis *Listener) error {
	//Handle listener setup
	s.mu.Lock()
	s.logf("serving")

	//TODO CASE HANDLING

	s.lis = lis

	s.mu.Unlock()
	//Begin Listen//accept loop
	//Sleep interval for null connect or previous connect
	var tempDelay time.Duration = 2
	for {
		newQueuePair, err := lis.Accept()

		if err != nil {
			//panic
			s.logf("Accept error: %v", err)
			return err
		}
		// Check null or against map to backoff
		if newQueuePair == nil {
			s.logf("null queue_pair: %v", newQueuePair)
			time.Sleep(tempDelay)
		}
		if _, ok := s.conns[newQueuePair.ClientId]; ok { //Queue pair already exists
			s.logf("already served queue_pair: %v", newQueuePair)
			// We are already servicing this queue
			time.Sleep(tempDelay)
		}

		//Start a new goroutine to deal with the new queuepair
		s.serveWG.Add(1)
		go func() {
			s.handleNewQueuePair(newQueuePair)
			s.serveWG.Done()
		}()
	}

}

// Fork a goroutine to handle just-accepted connection
func (s *Server) handleNewQueuePair(queuePair *QueuePair) {

	//Check for quit

	//Conn wrapper?

	//Deadline of inactivity?

	//Add to map
	s.conns[queuePair.ClientId] = queuePair

	//Launch dedicated thread to handle
	go func() {
		s.serveRequests()
		// If return from this method, connection has been closed
		// Remove and start servicing, close connection
	}()
}

func serveRequests() {

	// Call handle method as we read of queue appropriately.
}

func HandleMethod() {

}
func handleMethod() {

}

// Register Service, also gets generated by protoc, as Register(SERVICE NAME)Server
func (s *Server) RegisterService(desc *grpc.ServiceDesc, svr interface{}) {
	s.handlers.RegisterService(desc, svr)
}

func (s *Server) HandleMethods(svr interface{}) {
	s.unaryInterceptor = nil

	requestQueue := GetQueue(s.ShmQueueInfo.RequestShmaddr)
	responseQueue := GetQueue(s.ShmQueueInfo.ResponseShmaddr)

	for {

		message, err := consumeMessage(requestQueue)
		if err != nil {
			break
			//the channel has been shut down
		}

		messageTag := message.Header.Tag
		slice := message.Data[0:message.Header.Size]
		var message_req_meta ShmMessage
		if !sSerReqWritten {
			json.Unmarshal(slice, &message_req_meta)
			sSerReqStruct = message_req_meta
			sSerReqWritten = true
			if err != nil {
				// return err
				status.Errorf(codes.Unknown, "Codec Marshalling error: %s ", err.Error())
			}
		} else {
			message_req_meta = sSerReqStruct
		}

		//Parse bytes into object

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

		var resp_buffer []byte
		if !sserPayloadWritten {
			resp_buffer, err := codec.Marshal(resp)
			if err != nil {
				// return err
			}
			sserPayload = resp_buffer
			sserPayloadWritten = true
		}

		resp_buffer = sserPayload
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

		var serializedMessage []byte
		var data [600]byte
		if !sSerRespWritten {
			serializedMessage, err = json.Marshal(message_resp)
			sSerRespLen = copy(sSerRespData[:], serializedMessage)
			data = sSerRespData
			sSerRespWritten = true
			if err != nil {
				status.Errorf(codes.Unknown, "Codec Marshalling error: %s ", err.Error())
			}
		} else {
			data = sSerRespData
		}

		message_response := Message{
			Header: MessageHeader{
				Size: int32(sSerRespLen),
				Tag:  messageTag},
			Data: data,
		}

		produceMessage(responseQueue, message_response)

		if !NO_SERIALIZATION {
			sSerReqWritten = false
			sSerRespWritten = false
			sserPayloadWritten = false

		}

	}

}

// Shutdown the server
func (s *Server) Stop() {
	StopPollingQueue(s.requestQeuue)
	StopPollingQueue(s.responseQueue)

	defer Detach(s.ShmQueueInfo.RequestShmaddr)
	defer Detach(s.ShmQueueInfo.ResponseShmaddr)

	defer Remove(s.ShmQueueInfo.RequestShmid)
	defer Remove(s.ShmQueueInfo.ResponseShmid)
	// close(responseQueue.DetachQueue)

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

func (s *Server) logf(format string, args ...any) {
	if s.ErrorLog != nil {
		s.ErrorLog.Printf(format, args...)
	} else {
		log.Printf(format, args...)
	}
}
