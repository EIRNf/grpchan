package shmgrpc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"sync"

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
	BaseURL *url.URL
	//shm state info etc that might be needed
	ServiceName string
	//Connection metadata
	Metadata MessageMeta
	//Lock for concurrency
	Lock sync.Mutex
}

type MessageMeta struct {
	NumMessages int32
}

var _ grpc.ClientConnInterface = (*Channel)(nil)

var (
	cserReqData     [600]byte
	cserReqLen      int
	cserReqWritten  bool = false
	cserRespStruct  ShmMessage
	cserRespWritten bool = false

	cserPayload        []byte
	cserPayloadWritten bool = false
)

func NewChannel(url *url.URL, basePath string) *Channel {
	ch := new(Channel)
	ch.BaseURL = url

	//Initilize ShmQueueInfo
	requestKey, responseKey := GatherShmKeys(basePath)

	//Client -> Server Shm
	requestShmid, requestShmaddr := InitializeShmRegion(requestKey, Size, uintptr(ClientSegFlag))
	//Server -> Client Shm
	responseShmid, responseShmaddr := InitializeShmRegion(responseKey, Size, uintptr(ClientSegFlag))

	qi := QueueInfo{
		RequestShmid:    requestShmid,
		RequestShmaddr:  requestShmaddr,
		ResponseShmid:   responseShmid,
		ResponseShmaddr: responseShmaddr,
	}
	ch.ShmQueueInfo = &qi

	ch.Metadata = MessageMeta{
		NumMessages: 0,
	}

	ch.Lock = sync.Mutex{}

	log.Info().Msgf("New Channel: %v \n ", ch)
	log.Info().Msgf("New Channel RequestShmid: %v \n ", requestShmid)
	log.Info().Msgf("New Channel RespomseShmid: %v \n ", responseShmid)

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
	reqUrl := *ch.BaseURL
	reqUrl.Path = path.Join(reqUrl.Path, methodName)
	reqUrlStr := reqUrl.String()

	ctx, err := internal.ApplyPerRPCCreds(ctx, copts, fmt.Sprintf("shm:0%s", reqUrlStr), true)
	if err != nil {
		return err
	}

	codec := encoding.GetCodec(grpcproto.Name)

	if !cserPayloadWritten {
		serializedPayload, err := codec.Marshal(req)
		if err != nil {
			return err
		}
		cserPayload = serializedPayload
		cserPayloadWritten = true
	}

	// var reqPtr = unsafe

	messageRequest := &ShmMessage{
		Method:  methodName,
		Context: ctx,
		Headers: headersFromContext(ctx),
		Payload: ByteSlice2String(cserPayload),
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
	var data [600]byte
	if !cserReqWritten {
		serializedMessage, err = json.Marshal(messageRequest)
		cserReqLen = copy(cserReqData[:], serializedMessage)
		data = cserReqData
		cserReqWritten = true
		if err != nil {
			return err
		}
	} else {
		data = cserReqData
	}

	//START MESSAGING
	requestQueue := GetQueue(ch.ShmQueueInfo.RequestShmaddr)
	responseQueue := GetQueue(ch.ShmQueueInfo.ResponseShmaddr)

	message := Message{
		Header: MessageHeader{
			Size: int32(cserReqLen),
			Tag:  ch.Metadata.NumMessages},
		Data: data,
	}

	// pass into shared mem queue
	ch.Lock.Lock()
	produceMessage(requestQueue, message)
	ch.Lock.Unlock()

	//Receive Request
	ch.Lock.Lock()
	respMessage, err := consumeMessage(responseQueue)
	ch.Lock.Unlock()
	if err != nil {
		//This should hopefully not happen
		return err
	}

	if respMessage.Header.Tag != ch.Metadata.NumMessages {
		panic("Mismatched tag")
	}

	//Parse bytes into object
	slice := respMessage.Data[0:respMessage.Header.Size]
	var message_resp_meta ShmMessage
	if !cserRespWritten {
		json.Unmarshal(slice, &message_resp_meta)
		cserRespStruct = message_resp_meta
		cserRespWritten = true
		if err != nil {
			return err
		}
	} else {
		message_resp_meta = cserRespStruct
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

	if !NO_SERIALIZATION {
		cserReqWritten = false
		cserRespWritten = false
		cserPayloadWritten = false
	}

	// fmt.Printf("finished message num %d:", ch.Metadata.NumMessages)

	ret_err := codec.Unmarshal(payload, resp)
	return ret_err
}

func (ch *Channel) NewStream(ctx context.Context, desc *grpc.StreamDesc, methodName string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, nil
}
