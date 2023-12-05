package shmgrpc

import (
	"errors"
	"hash/fnv"
	"os"
	"syscall"
	"unsafe"

	"github.com/rs/zerolog/log"
)

const (
	RequestKey    = 1234  // Shared memory request key
	ResponseKey   = 1235  // Shared memory response key
	Size          = 16384 //+ 40 // Shared memory size
	Mode          = 0644  // Permissions for shared memory
	ServerSegFlag = IPC_CREAT | IPC_EXCL | Mode
	ClientSegFlag = IPC_CREAT | Mode
	MessageSize   = unsafe.Sizeof(Message{})
	QueueSize     = int32(Size) / int32(MessageSize)
)

// Hash name get key
func HashNameGetKey(input string) uintptr {
	h := fnv.New64a()
	h.Write([]byte(input))
	return uintptr(h.Sum64())
}

// Use a key value store to store keys, query for
func GatherShmKeys(fullServiceName string) (uintptr, uintptr) {
	//Option 1, use a key value store like redis

	//Option 2, use a hash of ServiceName, this is like not good for a few reasons.
	// Collisions "could" still occur, adding 1 makes this worse but I dont know enough cryptography to fully understand the consequeunces of my actions
	// This model falls apart in the face of multiple clients or multiple servers. It also falls apart at multiple instances
	// A proper solution is to implement something like grpc dialer which "i think" actually resovles for and sets up the
	// tcp connection by actually beginning a handshake with the given IP and Socket. This isn't really an option for
	// us but we need a way of dynamically creating regions of shared memory that can be easily and correctly resolved for
	// by a client

	//Convert the integer to a uintptr type
	requestKey := HashNameGetKey(fullServiceName)
	responseKey := HashNameGetKey(fullServiceName) + 1

	return requestKey, responseKey
}

func InitializeShmRegion(key, size, segFlag uintptr) (uintptr, uintptr) {

	log.Info().Msgf("New ShmRegion Key: %+v \n", key)
	log.Info().Msgf("New ShmRegion Size: %+v \n", size)
	log.Info().Msgf("New ShmRegion segFlag: %+v \n", segFlag)

	// Create a new shared memory segment
	shmid, _, errno := syscall.RawSyscall(syscall.SYS_SHMGET, key, size, segFlag)
	if errno != 0 {
		os.NewSyscallError("SYS_SHMGET", errno)
	}
	log.Info().Msgf("New ShmRegion shmid : %+v \n", shmid)
	log.Info().Msgf("New ShmRegion errno : %+v \n", errno)

	shmaddr, _, errno := syscall.RawSyscall(syscall.SYS_SHMAT, shmid, uintptr(0), segFlag)
	if errno != 0 {
		os.NewSyscallError("SYS_SHMAT", errno)
	}
	log.Info().Msgf("New ShmRegion shmaddr : %+v \n", shmaddr)
	log.Info().Msgf("New ShmRegion errno : %+v \n", errno)

	return shmid, shmaddr
}

func AttachToShmRegion(shmid, segFlag uintptr) uintptr {

	shmaddr, _, errno := syscall.RawSyscall(syscall.SYS_SHMAT, shmid, uintptr(0), segFlag)
	if errno != 0 {
		os.NewSyscallError("SYS_SHMAT", errno)
	}

	return shmaddr
}

func Remove(shm_id uintptr) error {
	_, _, errno := syscall.Syscall(syscall.SYS_SHMCTL, shm_id, 0, 0)
	if errno != 0 {
		return errors.New(errno.Error())
	}
	return nil
}

// Detach used to detach from memory segment
func Detach(shmaddr uintptr) error {
	_, _, errno := syscall.Syscall(syscall.SYS_SHMDT, shmaddr, 0, 0)
	if errno != 0 {
		return errors.New(errno.Error())
	}
	return nil
}
