package shmgrpc

import (
	"bytes"
	"context"
	"net/url"
	"sync"
	"time"

	"github.com/fullstorydev/grpchan/internal"

	jsoniter "github.com/json-iterator/go"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/encoding"
	grpcproto "google.golang.org/grpc/encoding/proto"
	"google.golang.org/grpc/status"

	"github.com/rs/zerolog/log"
)

// All required info for a client to communicate with a server
type Channel struct {
	ShmQueueInfo *QueueInfo
	//URL of endpoint (might be useful in the future)
	targetAddress *url.URL
	sourceAddress string
	//shm state info etc that might be needed

	queuePair *QueuePair
	//Connection metadata
	Metadata MessageMeta
	//Lock for concurrency
	Lock sync.Mutex

	responseBuffer *bytes.Buffer

	prev_time       time.Time
	cresponseBuffer []byte
}

func (ch *Channel) timestamp_dif() string {
	// if s.prev_time != nil {
	// 	s.prev_time = time.Now()

	// }
	dif := time.Since(ch.prev_time).String()
	ch.prev_time = time.Now()
	return dif
}

type MessageMeta struct {
	NumMessages int32
}

var _ grpc.ClientConnInterface = (*Channel)(nil)

func NewChannel(sourceAddress string, targetAddress string) *Channel {
	ch := new(Channel)
	ch.sourceAddress = targetAddress
	ch.targetAddress, _ = url.Parse(targetAddress)

	log.Info().Msgf("Client: Opening New Channel \n")
	ch.queuePair = ClientOpen(sourceAddress, targetAddress, 512)

	if ch.queuePair == nil {
		log.Info().Msgf("error establishing notnets conn: %v \n ", ch)
	}

	ch.Metadata = MessageMeta{
		NumMessages: 0,
	}

	ch.Lock = sync.Mutex{}

	log.Info().Msgf("Client: New Channel: %v \n ", ch.queuePair.ClientId)
	log.Info().Msgf("Client: New Channel RequestShmid: %v \n ", ch.queuePair.RequestShmaddr)
	log.Info().Msgf("Client: New Channel RespomseShmid: %v \n ", ch.queuePair.ResponseShmaddr)

	ch.responseBuffer = bytes.NewBuffer(nil)
	ch.cresponseBuffer = make([]byte, 512)

	return ch
}

func (ch *Channel) incrementNumMessages() {
	//We can wrap this in a lock if necessary
	ch.Metadata.NumMessages += 1
}

func (ch *Channel) Invoke(ctx context.Context, methodName string, req, resp interface{}, opts ...grpc.CallOption) error {

	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	// jsoniter.RegisterTypeDecoderFunc("context", func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
	// 	s := iter.ReadString()
	// 	if len(s) == 0 { //Empty string
	// 		*((*context.Context)(ptr)) = nil
	// 	} else {
	// 		*((*context.Context)(ptr)) = jsoniter.op
	// 	}

	// })

	ch.prev_time = time.Now()

	// log.Info().Msgf("Client Invoke: %v ", req)

	//Get Call Options for
	copts := internal.GetCallOptions(opts)

	// var copts internal.CallOptions
	// for _, o := range opts {
	// 	switch o := o.(type) {
	// 	case grpc.HeaderCallOption:
	// 		copts.Headers = append(copts.Headers, o.HeaderAddr)
	// 	case grpc.TrailerCallOption:
	// 		copts.Trailers = append(copts.Trailers, o.TrailerAddr)
	// 	case grpc.PeerCallOption:
	// 		copts.Peer = append(copts.Peer, o.PeerAddr)
	// 	case grpc.PerRPCCredsCallOption:
	// 		copts.Creds = o.Creds
	// 	case grpc.MaxRecvMsgSizeCallOption:
	// 		copts.MaxRecv = o.MaxRecvMsgSize
	// 	case grpc.MaxSendMsgSizeCallOption:
	// 		copts.MaxSend = o.MaxSendMsgSize
	// 	}
	// }

	// log.Info().Msgf("call options: %s", ch.timestamp_dif())

	// Get headersFromContext
	// reqUrl := ch.targetAddress
	// reqUrl.Path = path.Join(reqUrl.Path, methodName)
	// reqUrlStr := reqUrl.String()

	// // log.Info().Msgf("path handling: %s", ch.timestamp_dif())

	// ctx, err := internal.ApplyPerRPCCreds(ctx, copts, fmt.Sprintf("shm:0%s", reqUrlStr), true)
	// if err != nil {
	// 	return err
	// }

	// log.Info().Msgf("appply creds: %s", ch.timestamp_dif())

	codec := encoding.GetCodec(grpcproto.Name)

	serializedPayload, err := codec.Marshal(req)
	if err != nil {
		return err
	}

	// log.Info().Msgf("codec marshal: %s", ch.timestamp_dif())

	messageRequest := &ShmMessage{
		Method:  methodName,
		ctx:     ctx,
		Headers: headersFromContext(ctx),
		// Trailers: trailersFrom,
		Payload: ByteSlice2String(serializedPayload),
	}

	// Create a fixed-length byte array
	// var byteArray [unsafe.Sizeof(messageRequest)]byte

	// Copy the bytes of the struct into the byte array
	// messageRequestBytes := *(*[unsafe.Sizeof(messageRequest)]byte)(unsafe.Poier(&messageRequest))
	// messageRequestBytes := fmt.Sprintf("%+v\n", messageRequest)
	// copy(byteArray[:], messageRequestBytes[:])

	// we have the meta request
	// Marshall to build rest of system

	var serializedMessage []byte
	serializedMessage, err = json.Marshal(messageRequest)

	// messageRequest = nil
	if err != nil {
		return err
	}
	// log.Info().Msgf("marshal: %s", ch.timestamp_dif())

	//START MESSAGING
	// pass into shared mem queue
	// ch.Lock.Lock()
	// log.Info().Msgf("Client: Message Sent: %v \n ", serializedMessage)
	// runtime.LockOSThread()
	ch.queuePair.ClientSendRpc(serializedMessage, len(serializedMessage))
	// serializedMessage = nil
	// ch.Lock.Unlock()

	// log.Info().Msgf("send request: %s", ch.timestamp_dif())

	//Receive Request
	// ch.Lock.Lock()
	// b := bytes.NewBuffer(nil)
	// buf := make([]byte, 512)
	//iterate and append to dynamically allocated data until all data is read
	var size int
	for {
		size = ch.queuePair.ClientReceiveBuf(ch.cresponseBuffer, len(ch.cresponseBuffer))
		// log.Info().Msgf("Client: Reads: %v", buf)

		ch.responseBuffer.Write(ch.cresponseBuffer)
		if size == 0 { //Have full payload
			break
		}
	} // ch.Lock.Unlock()
	// log.Info().Msgf("receive response: %s", ch.timestamp_dif())

	// runtime.UnlockOSThread()

	// log.Info().Msgf("Client: Message Received: %v \n ", b.String())

	var message_resp_meta ShmMessage

	// json.Unmarshal(b.Bytes(), &message_resp_meta)

	// fmt.Printf("resp message: %v", ch.responseBuffer.String())

	// err = ffjson.NewDecoder().DecodeReader(ch.responseBuffer, &message_resp_meta)
	dec := json.NewDecoder(ch.responseBuffer)
	err = dec.Decode(&message_resp_meta)

	// decoder := json.NewDecoder(&io.LimitedReader{N: 512, R: ch.responseBuffer})
	// decoder.Decode(&message_resp_meta)
	// decoder = nil
	// json.Unmarshal(bytes.Trim(b, "\x00"))

	// log.Info().Msgf("decode: %s", ch.timestamp_dif())

	if err != nil {
		return err // TODO BAD
	}

	// payload := unsafeGetBytes(message_resp_meta.Payload)
	payload := []byte(message_resp_meta.Payload)

	copts.SetHeaders(message_resp_meta.Headers)
	copts.SetTrailers(message_resp_meta.Trailers)

	// ipc.Msgctl(qid, ipc.IPC_RMID)
	// var ret_err error
	// if !cserPayloadRespWritten {
	// copy(cserPayloadResp, resp)

	// cserPayloadRespWritten = true
	// }
	// resp = cserPayloadResp

	//Update total number of back and forth messages
	ch.incrementNumMessages()
	// fmt.Printf("finished message num %d:", ch.Metadata.NumMessages)

	// we fire up a goroutine to read the response so that we can properly
	// respect any context deadline (e.g. don't want to be blocked, reading
	// from socket, long past requested timeout).
	// respCh := make(chan struct{})
	// select {
	// case <-ctx.Done():
	// 	return statusFromContextError(ctx.Err())
	// case <-respCh:
	// }
	// if err != nil {
	// 	return err
	// }

	ret_err := codec.Unmarshal(payload, resp)
	// log.Info().Msgf("unmarshal: %s", ch.ztimestamp_dif())

	return ret_err
}

// statusFromContextError translates the given error, returned by a call to
// context.Context.Err(), into a suitable GRPC error. If the given error is
// not a context error (e.g. neither deadline exceeded nor canceled) then it
// is returned as is.
func statusFromContextError(err error) error {
	if err == context.DeadlineExceeded {
		return status.Error(codes.DeadlineExceeded, err.Error())
	} else if err == context.Canceled {
		return status.Error(codes.Canceled, err.Error())
	}
	return err
}

func (ch *Channel) NewStream(ctx context.Context, desc *grpc.StreamDesc, methodName string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, nil
}
