package shm

import (
	"encoding/binary"
	"fmt"
)

const (
	headerLen = 2
	bodyLen   = 4
)

// BinaryMessage describes binary message
type binaryMessage struct {
	LogType uint16
	Len     int
	Body    []byte
}

func newBinaryMessage(logType uint16, data []byte) *binaryMessage {
	return &binaryMessage{
		LogType: logType,
		Len:     len(data),
		Body:    data,
	}
}

// Serialize is to add log type & body length
func (b *binaryMessage) serialize() []byte {
	buf := make([]byte, b.Len+headerLen+bodyLen)
	binary.BigEndian.PutUint16(buf[:headerLen], b.LogType)
	binary.BigEndian.PutUint32(buf[headerLen:bodyLen+headerLen], uint32(b.Len))
	copy(buf[bodyLen+headerLen:], b.Body)
	return buf
}

// Deserialize turns bytes to BinaryMessage
func deserialize(data []byte) (*binaryMessage, error) {
	// data size must be greater than 2 bytes
	if len(data) < headerLen+bodyLen {
		return nil, fmt.Errorf("too short data size")
	}

	message := &binaryMessage{}
	message.LogType = binary.BigEndian.Uint16(data[:headerLen])
	message.Len = int(binary.BigEndian.Uint32(data[headerLen : headerLen+bodyLen]))
	if (message.Len + headerLen + bodyLen) != len(data) {
		return nil, fmt.Errorf("mismatch body length")
	}
	message.Body = data[headerLen+bodyLen:]
	return message, nil
}
