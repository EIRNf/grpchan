package shmgrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"path"
	"sync"
	"time"

	"github.com/fullstorydev/grpchan/internal"

	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
	grpcproto "google.golang.org/grpc/encoding/proto"

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
}

type MessageMeta struct {
	NumMessages int32
}

var _ grpc.ClientConnInterface = (*Channel)(nil)

func NewChannel(sourceAddress string, targetAddress string) *Channel {
	ch := new(Channel)
	ch.sourceAddress = targetAddress
	ch.targetAddress, _ = url.Parse(targetAddress)

	time.Sleep(2 * time.Second)
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

	return ch
}

func (ch *Channel) incrementNumMessages() {
	//We can wrap this in a lock if necessary
	ch.Metadata.NumMessages += 1
}

func (ch *Channel) Invoke(ctx context.Context, methodName string, req, resp interface{}, opts ...grpc.CallOption) error {

	// log.Info().Msgf("Client Invoke: %v ", req)

	//Get Call Options for
	copts := internal.GetCallOptions(opts)

	//Get headersFromContext
	reqUrl := ch.targetAddress
	reqUrl.Path = path.Join(reqUrl.Path, methodName)
	reqUrlStr := reqUrl.String()

	ctx, err := internal.ApplyPerRPCCreds(ctx, copts, fmt.Sprintf("shm:0%s", reqUrlStr), true)
	if err != nil {
		return err
	}

	codec := encoding.GetCodec(grpcproto.Name)

	serializedPayload, err := codec.Marshal(req)
	if err != nil {
		return err
	}

	messageRequest := &ShmMessage{
		Method:  methodName,
		Context: ctx,
		Headers: headersFromContext(ctx),
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

	if err != nil {
		return err
	}

	//START MESSAGING
	// pass into shared mem queue
	// ch.Lock.Lock()
	log.Info().Msgf("Client: Message Sent: %v \n ", serializedMessage)
	ch.queuePair.ClientSendRpc(serializedMessage, len(serializedMessage))
	// ch.Lock.Unlock()

	//Receive Request
	// ch.Lock.Lock()
	b := bytes.NewBuffer(nil)
	buf := make([]byte, 512)
	//iterate and append to dynamically allocated data until all data is read
	var size int
	for {
		size = ch.queuePair.ClientReceiveBuf(buf, len(buf))
		log.Info().Msgf("Client: Reads: %v", buf)

		b.Write(buf)
		if size == 0 { //Have full payload
			break
		}
	} // ch.Lock.Unlock()

	log.Info().Msgf("Client: Message Received: %v \n ", b.String())

	var message_resp_meta ShmMessage
	// json.Unmarshal(b.Bytes(), &message_resp_meta)
	json.NewDecoder(&io.LimitedReader{N: 512, R: b}).Decode(&message_resp_meta)

	if err != nil {
		return err
	}

	payload := unsafeGetBytes(message_resp_meta.Payload)

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

	ret_err := codec.Unmarshal(payload, resp)
	return ret_err
}

func (ch *Channel) NewStream(ctx context.Context, desc *grpc.StreamDesc, methodName string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, nil
}
