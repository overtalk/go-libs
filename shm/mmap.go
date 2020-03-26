package shm

import "errors"

var errOutOfCapacity = errors.New("out of capacity")

type tag struct {
	ReadIndex  int32
	WriteIndex int32
}

// Mem defines the share mem of mMap
type mMap struct {
	*tag
	queue []byte
}

func (m *mMap) save(buf []byte) error {
	size := int32(cap(m.queue))

	currentHead := m.ReadIndex
	var capacity int32
	length := len(buf)

	if currentHead > m.WriteIndex {
		capacity = currentHead - m.WriteIndex
	} else {
		capacity = size - (m.WriteIndex - currentHead)
	}

	// above capacity
	if int(capacity) < length {
		return errOutOfCapacity
	}

	for index, bit := range buf {
		m.queue[(m.WriteIndex+int32(index))%size] = bit
	}

	m.WriteIndex = (m.WriteIndex + int32(length)) % size
	return nil
}

func (m *mMap) get() []byte {
	size := cap(m.queue)
	currentTail := m.WriteIndex
	var ret []byte

	switch {
	case m.ReadIndex == currentTail:
		return nil
	case m.ReadIndex < currentTail:
		ret = append(ret, m.queue[m.ReadIndex:currentTail]...)
	default:
		ret = append(ret, m.queue[m.ReadIndex:size]...)
		ret = append(ret, m.queue[0:currentTail]...)
	}

	return ret
}

func (m *mMap) release(currentTail int32) {
	m.ReadIndex = currentTail
}

// Flush clear all the data
func (m *mMap) flush() {
	m.ReadIndex = 0
	m.WriteIndex = 0
}
