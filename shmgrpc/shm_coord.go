package shmgrpc

import (
	"time"
	"unsafe"
)

// Common structs
type CoordHeader struct {
	Counter        int
	lock           *FileMutex
	AvailableSlots [availableCoordSlots]ReservePair // This needs to be thread safe
	CoordSlots     [availableCoordSlots]CoordRow
}

type ReservePair struct {
	ClientRequest bool
	ServerHandled bool
}

type CoordRow struct {
	//Might need a lock here
	ClientID       uint //Only written to by Client, how do we even know id?
	ShmCreated     bool //Only written to by Server //If false server creates it and adds keys to
	RequestShmKey  uint //Only written to by Server // if -1 then that means it hasnt been created?
	ResponseShmKey uint //Only written to by Server
	Detach         bool //Only written to by Client
}

const (
	// coordRegionSize     uint = 1648
	availableCoordSlots = 10
)

// Client Section
func getQueuePair(header *CoordHeader) (RequestShmKey, ResponseShmKey uint) {

	reservedSlot := -1
	for {
		//Reserve a slot, if not available wait and try again
		header.lock.RLock()
		for i := 0; i < availableCoordSlots; i++ {
			if !header.AvailableSlots[i].ClientRequest { // header.AvailableSlots[i].ClientRequest == false && header.AvailableSlots[i].ServerHandled == false
				reservedSlot = i

				//Place flag in critical section to request handling
				header.AvailableSlots[reservedSlot] = ReservePair{
					true,
					false,
				}

				//Place Coord Request
				header.CoordSlots[reservedSlot] = CoordRow{
					0,
					false,
					0,
					0,
					false,
				}
				break
			}
		}
		header.lock.RUnlock()

		if reservedSlot != -1 {
			break
		}
	}

	//Wait until the clost has been handled, check over and over again with some time
	for {
		//Handled
		if header.CoordSlots[reservedSlot].ShmCreated {
			return header.CoordSlots[reservedSlot].RequestShmKey, header.CoordSlots[reservedSlot].ResponseShmKey
		}
		time.Sleep(5)
	}

}

// Server Section
func HandleIncomingShmRequests(server Server, header *CoordHeader) {
	//Get access to the critical section

	//Launch go routine to iterate though header until an entry requesting a connection is found
	for {
		//lock access to critical section array
		header.lock.RLock()
		for i := 0; i < availableCoordSlots; i++ {
			switch {
			case header.AvailableSlots[i].ClientRequest == true && header.AvailableSlots[i].ServerHandled == false:
				//New client needs handling
				AssignClientQueue(i)

			default:
				continue
			}
		}
		header.lock.RUnlock()

	}

}

func AssignClientQueue(slot int) {

	//Create Queues

}

// Create and instantiate coord shm, used by server to create coord
func CreateCoord(serviceName string) *CoordHeader {
	coordKey := HashNameGetKey(serviceName)

	//Create lock
	lock, _ := New("coord")

	CoordHeader := CoordHeader{
		0,
		lock,
		[availableCoordSlots]ReservePair{},
		[availableCoordSlots]CoordRow{},
	}

	//initialize shared memory region
	_, coordsmhaddr := InitializeShmRegion(coordKey, unsafe.Sizeof(CoordHeader), uintptr(ServerSegFlag))

	coordPtr := GetCoordPtr(coordsmhaddr)
	*coordPtr = CoordHeader

	return coordPtr
}

// Used by client to attach to existing service
func AttachToCoord(serviceName string) *CoordHeader {
	coordKey := HashNameGetKey(serviceName)

	//Create lock
	lock, _ := New("coord")

	//Only used for size calculation no assignment of coordPtr
	CoordHeader := CoordHeader{
		0,
		lock,
		[availableCoordSlots]ReservePair{},
		[availableCoordSlots]CoordRow{},
	}

	//initialize shared memory region
	_, coordsmhaddr := InitializeShmRegion(coordKey, unsafe.Sizeof(CoordHeader), uintptr(ServerSegFlag))

	return GetCoordPtr(coordsmhaddr)
}

func GetCoordPtr(shmaddr uintptr) *CoordHeader {
	coordPtr := (*CoordHeader)(unsafe.Pointer(shmaddr)) //TODO: this is correct actually
	// fmt.Printf("unsafeGetBytes pointer: %p\n", &queuePtr)
	return coordPtr
}

// Poll coord shm
