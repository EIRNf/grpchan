package shmgrpc_test

import (
	"bytes"
	"testing"

	"github.com/fullstorydev/grpchan/shmgrpc"
)

// TestHelloName calls greetings.Hello with a name, checking
// for a valid return value.
func Test_RPC(t *testing.T) {

	notnets_context := shmgrpc.RegisterServer("hello")

	queuePair := shmgrpc.ClientOpen("sourceAddress", "hello", 512)

	notnets_context.Accept()

	mes := []byte{12, 14, 12, 32}

	queuePair.ClientSendRpc(mes, len(mes))

	bs := bytes.NewBuffer(nil)
	sbuf := make([]byte, 512)
	//iterate and append to dynamically allocated data until all data is read
	var size int
	for {
		size = queuePair.ServerReceiveBuf(sbuf, len(sbuf))
		bs.Write(sbuf)
		if size == 0 { //Have full payload
			break
		}
	}

	queuePair.ServerSendRpc(bs.Bytes(), len(bs.Bytes()))

	bc := bytes.NewBuffer(nil)
	cbuf := make([]byte, 512)
	//iterate and append to dynamically allocated data until all data is read
	for {
		size = queuePair.ClientReceiveBuf(cbuf, len(cbuf))
		bc.Write(cbuf)
		if size == 0 { //Have full payload
			break
		}
	} // ch.Lock.Unlock()

	print(bc.Bytes())

}
