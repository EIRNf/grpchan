//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package shmgrpc

import (
	"context"
	"errors"
	"reflect"
	"syscall"
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
	NO_SERIALIZATION bool = true

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

	//OPEN
	O_CREAT  = 0x40
	O_RDONLY = 0x0

	//LOCK
	LOCK_SH = 0x1
	LOCK_EX = 0x2
	LOCK_NB = 0x4
	LOCK_UN = 0x8
)

const (
	S_IRUSR = 0400         /* Read by owner.  */
	S_IWUSR = 0200         /* Write by owner.  */
	S_IRGRP = S_IRUSR >> 3 /* Read by group.  */
	S_IWGRP = S_IWUSR >> 3 /* Write by group.  */
)

type ShmMessage struct {
	Method  string          `json:"method"`
	Context context.Context `json:"context"`
	// Headers  map[string][]byte `json:"headers,omitempty"`
	// Trailers map[string][]byte `json:"trailers,omitempty"`
	Headers  metadata.MD `json:"headers"`
	Trailers metadata.MD `json:"trailers"`
	Payload  string      `json:"payload"`
	// Payload interface{}     `protobuf:"bytes,3,opt,name=method,proto3" json:"payload"`
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

const (
	//permenant directory flags
	mkdirPerm = 0750
)

var AlreadyLocked = errors.New("lock already acquired")

// FileMutex is similar to sync.RWMutex, but also synchronizes across processes.
// This implementation is based on flock syscall.
type FileMutex struct {
	fd int
}

func New(filename string) (*FileMutex, error) {

	fd, err := syscall.Open(filename, O_CREAT|O_RDONLY, mkdirPerm)
	if err != nil {
		return nil, err
	}
	return &FileMutex{fd: fd}, nil
}
func (m *FileMutex) Lock() error {
	return syscall.Flock(m.fd, LOCK_EX)
}

func (m *FileMutex) TryLock() error {

	if err := syscall.Flock(m.fd, LOCK_EX|LOCK_NB); err != nil {
		if errno, ok := err.(syscall.Errno); ok {
			if errno == syscall.Errno(0xb) {
				return AlreadyLocked
			}
		}
		return err
	}
	return nil
}

func (m *FileMutex) Unlock() error {
	return syscall.Flock(m.fd, LOCK_UN)
}

func (m *FileMutex) RLock() error {
	return syscall.Flock(m.fd, LOCK_SH)
}

func (m *FileMutex) RUnlock() error {
	return syscall.Flock(m.fd, LOCK_UN)
}

// Close unlocks the lock and closes the underlying file descriptor.
func (m *FileMutex) Close() error {
	if err := syscall.Flock(m.fd, LOCK_UN); err != nil {
		return err
	}
	return syscall.Close(m.fd)
}
