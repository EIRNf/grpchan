package shmgrpc

import (
	"errors"
	"unsafe"
)

// var produce_lock sync.Mutex
// var consume_lock sync.Mutex

type MessageHeader struct {
	Size int32
	Tag  int32
}

type Message struct {
	Header MessageHeader
	Data   [600]byte // Maximum payload size
}

type Queue struct {
	// ProducerFlag bool
	// ConsumerFlag bool
	// mu          sync.Mutex
	Head        int32
	Tail        int32
	Count       int32
	TotalCount  int32
	BufferSize  int32
	StopPolling bool // DetachQueue chan bool
}

func initializeQueue(shmaddr uintptr) *Queue {
	// Initialize the circular buffer structure
	queue := Queue{
		Head:        0,
		Tail:        0,
		BufferSize:  QueueSize,
		Count:       0,
		TotalCount:  0,
		StopPolling: false,
	}
	// fmt.Printf("Queue size: %d\n", unsafe.Sizeof(queue))
	queuePtr := GetQueue(shmaddr)
	*queuePtr = queue
	return queuePtr
}

func StopPollingQueue(queuePtr *Queue) {
	queuePtr.StopPolling = true
}

func produceMessage(queuePtr *Queue, message Message) {
	// produce_lock.Lock()

	// for isFull(queuePtr) {
	// 	// Wait for space to become available
	// }

	// Wait until there's space in the circular buffer
poll:
	for {
		switch {
		case queuePtr.StopPolling:
			return
		default:
			// Wait for space to become available
			if isFull(queuePtr) {
				// time.Sleep(time.Nanosecond)
				continue
			}
			break poll
		}

	}

	// Enqueue the message into the circular buffer
	enqueue(queuePtr, &message)
	// fmt.Printf("Producer: Message enqueued (Size: %s)\n", message.Data)
	// produce_lock.Unlock()

}

func consumeMessage(queuePtr *Queue) (Message, error) {
	// consume_lock.Lock()
	var message Message

poll:
	for {
		switch {
		case queuePtr.StopPolling:
			//This might be problematic
			message = Message{}
			return message, errors.New("SharedMem detached")
		default:
			// Wait for space to become available
			if isEmpty(queuePtr) {
				// time.Sleep(time.Nanosecond)
				continue
			} else {
				// Dequeue the message from the circular buffer
				message = dequeue(queuePtr)
				break poll
			}
		}
		// Wait for a message to become available
	}
	// consume_lock.Unlock()
	return message, nil

	// fmt.Printf("Consumer: Received message (Size: %d): %s\n", message.Header.Size, string(message.Data[:message.Header.Size]))
}

func GetQueue(shmaddr uintptr) *Queue {
	queuePtr := (*Queue)(unsafe.Pointer(shmaddr)) //TODO: this is correct actually
	// fmt.Printf("unsafeGetBytes pointer: %p\n", &queuePtr)
	return queuePtr
}

func isFull(queue *Queue) bool {
	// queue.mu.Lock()
	isFull := (queue.Tail+1)%queue.BufferSize == queue.Head
	// queue.mu.Unlock()
	return isFull
}

func isEmpty(queue *Queue) bool {
	// queue.mu.Lock()
	isEmpty := queue.Head == queue.Tail
	// queue.mu.Unlock()
	return isEmpty
}

func enqueue(queue *Queue, message *Message) {
	// queue.mu.Lock()
	messageArray := (*[QueueSize]Message)(unsafe.Pointer(uintptr(unsafe.Pointer(queue)) + unsafe.Sizeof(*queue)))
	messageArray[queue.Tail] = *message
	queue.Count++
	queue.TotalCount++
	// fmt.Printf("Enqueue Count: %d\n", queue.Count)
	// fmt.Printf("Total Count: %d\n", queue.TotalCount)
	// fmt.Printf("Array pointer: %p\n", &messageArray[queue.Tail])
	// fmt.Printf("Array pos: %d\n", messageArray[queue.Tail])
	queue.Tail = (queue.Tail + 1) % queue.BufferSize
	// queue.mu.Unlock()
}

func dequeue(queue *Queue) Message {
	// queue.mu.Lock()
	message := (*[QueueSize]Message)(unsafe.Pointer(uintptr(unsafe.Pointer(queue)) + unsafe.Sizeof(*queue)))[queue.Head]
	queue.Head = (queue.Head + 1) % queue.BufferSize
	queue.Count--
	// queue.mu.Unlock()
	// fmt.Printf("Dequeue Count: %d\n", queue.Count)
	return message
}
