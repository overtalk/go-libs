package shm

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// CacheItem defines the item store in mMap cache
type CacheItem struct {
	ProtoID    uint16 `json:"id"`
	Key        string `json:"ukey"`
	ProtoBytes []byte `json:"data"`
}

// Serialize define the func to convert item to []byte
func (c *CacheItem) Serialize() []byte {
	keyBytes := []byte(c.Key)
	buf := make([]byte, 2+len(keyBytes)+len(c.ProtoBytes)+4)
	binary.BigEndian.PutUint16(buf[:2], c.ProtoID)
	binary.BigEndian.PutUint32(buf[2:6], uint32(len(keyBytes)))
	copy(buf[6:6+len(keyBytes)], keyBytes)
	copy(buf[6+len(keyBytes):], c.ProtoBytes)
	return buf
}

// Deserialize define the func to convert item from []byte
func (c *CacheItem) Deserialize(data []byte) error {
	if len(data) < 6 {
		return fmt.Errorf("too short data size")
	}
	c.ProtoID = binary.BigEndian.Uint16(data[:2])
	keyBytesLen := int(binary.BigEndian.Uint32(data[2:6]))
	if keyBytesLen+6 >= len(data) {
		return fmt.Errorf("mismatch body length")
	}
	c.Key = string(data[6 : 6+keyBytesLen])
	c.ProtoBytes = data[6+keyBytesLen:]
	return nil
}

// Serialize convert *CacheItem/[]*CacheItem to []byte
func Serialize(in interface{}) ([]byte, error) {
	var (
		data []byte
		err  error
	)

	switch in.(type) {
	case *CacheItem:
		data = newBinaryMessage(0, in.(*CacheItem).Serialize()).serialize()
	case []*CacheItem:
		for _, item := range in.([]*CacheItem) {
			data = append(data, newBinaryMessage(0, item.Serialize()).serialize()...)
		}
	default:
		err = errors.New("invalid type")
	}

	return data, err
}

// Deserialize convert []byte to []*CacheItem
func Deserialize(data []byte) (interface{}, error) {
	baseIndex := 0
	reader := bytes.NewReader(data)
	var messages [][]byte
	var ret []*CacheItem
	for {
		// get the header
		header := make([]byte, headerLen+bodyLen)
		if _, err := io.ReadFull(reader, header); err != nil {
			break
		}

		length := int(binary.BigEndian.Uint32(data[baseIndex+headerLen : baseIndex+headerLen+bodyLen]))

		// get the body
		bodyByte := make([]byte, length)
		if _, err := io.ReadFull(reader, bodyByte); err != nil {
			break
		}

		// send to message Chan
		messages = append(messages, bodyByte)
		baseIndex += (length + 6)
	}

	for _, v := range messages {
		temp := &CacheItem{}
		if err := temp.Deserialize(v); err != nil {
			return nil, err
		}

		ret = append(ret, temp)
	}

	return ret, nil
}
