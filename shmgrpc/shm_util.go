package shmgrpc

import (
	"context"
	"reflect"
	"unsafe"

	"google.golang.org/grpc/metadata"
)

type QueueInfo struct {
	RequestShmid    uintptr
	RequestShmaddr  uintptr
	ResponseShmid   uintptr
	ResponseShmaddr uintptr
}

type Flag int

// https://github.com/torvalds/linux/blob/master/include/uapi/linux/ipc.h
const (
	// Disable Serialization
	NO_SERIALIZATION bool = false

	/* resource get request flags */
	IPC_CREAT  Flag = 00001000 /* create if key is nonexistent */
	IPC_EXCL   Flag = 00002000 /* fail if key exists */
	IPC_NOWAIT Flag = 00004000 /* return error on wait */

	/* Permission flag for shmget.  */
	SHM_R Flag = 0400 /* or S_IRUGO from <linux/stat.h> */
	SHM_W Flag = 0200 /* or S_IWUGO from <linux/stat.h> */

	/* Flags for `shmat'.  */
	SHM_RDONLY Flag = 010000 /* attach read-only else read-write */
	SHM_RND    Flag = 020000 /* round attach address to SHMLBA */

	/* Commands for `shmctl'.  */
	SHM_REMAP Flag = 040000  /* take-over region on attach */
	SHM_EXEC  Flag = 0100000 /* execution access */

	SHM_LOCK   Flag = 11 /* lock segment (root only) */
	SHM_UNLOCK Flag = 12 /* unlock segment (root only) */
)

const (
	S_IRUSR = 0400         /* Read by owner.  */
	S_IWUSR = 0200         /* Write by owner.  */
	S_IRGRP = S_IRUSR >> 3 /* Read by group.  */
	S_IWGRP = S_IWUSR >> 3 /* Write by group.  */
)

type ShmMessage struct {
	Method string          `json:"method"`
	ctx    context.Context `json:"omitempty"`
	// Headers  map[string][]byte `json:"headers,omitempty"`
	// Trailers map[string][]byte `json:"trailers,omitempty"`
	Headers  metadata.MD `json:"headers"`
	Trailers metadata.MD `json:"trailers"`
	Payload  string      `json:"payload"`
	// Payload interface{}     `protobuf:"bytes,3,opt,name=method,proto3" json:"payload"`
}

// Gets background context
// For outgoing client requests, the context controls cancellation.
//
// For incoming server requests, the context is canceled when the
// client's connection closes, the request is canceled (with HTTP/2),
// or when the ServeHTTP method returns.
func (mes *ShmMessage) Context(ctx context.Context) context.Context {
	if ctx != nil {
		return ctx
	}
	return context.Background()
}

// WithContext returns a shallow copy of r with its context changed
// to ctx. The provided ctx must be non-nil.
func (mes *ShmMessage) WithContext(ctx context.Context) *ShmMessage {
	if ctx == nil {
		panic("nil context")
	}
	mes2 := new(ShmMessage)
	*mes2 = *mes
	mes2.ctx = ctx
	return mes2
}

// headersFromContext returns HTTP request headers to send to the remote host
// based on the specified context. GRPC clients store outgoing metadata into the
// context, which is translated into headers. Also, a context deadline will be
// propagated to the server via GRPC timeout metadata.
func headersFromContext(ctx context.Context) metadata.MD {
	md, ok := metadata.FromOutgoingContext(ctx)
	if ok {
		//great
	}

	return md
}

func unsafeGetBytes(s string) []byte {
	// fmt.Printf("unsafeGetBytes pointer: %p\n", unsafe.Pointer((*reflect.StringHeader)(unsafe.Pointer(&s)).Data))
	return (*[0x7fff0000]byte)(unsafe.Pointer(
		(*reflect.StringHeader)(unsafe.Pointer(&s)).Data),
	)[:len(s):len(s)]
}

func ByteSlice2String(bs []byte) string {
	// fmt.Printf("ByteSlice2String pointer: %p\n", unsafe.Pointer(&bs))
	return *(*string)(unsafe.Pointer(&bs))
}
