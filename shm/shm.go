package shm

import (
	"errors"
	"os"
	"sync/atomic"
	"syscall"
	"unsafe"
)

const (
	maxCapacity uint64 = 10 << 30 // 10G
	headSize           = 28
)

// head struct `size +currentBlockIndex + tag + tag + data`
type SHM struct {
	currentBlockIndex *int32
	size              *uint64
	mMaps             [2]*mMap
}

func NewShm(path string, size uint64) (*SHM, error) {
	if size > maxCapacity {
		return nil, errOutOfCapacity
	}

	fd, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return nil, err
	}

	info, err := fd.Stat()
	if err != nil {
		return nil, err
	}

	// mmap不会更改底层文件的大小，我们要确保访问的映射地址不会超过文件大小，否则会panic
	// 这里设置一下底层文件大小
	if info.Size() != int64(size*2+headSize) {
		if err := fd.Truncate(int64(size*2 + headSize)); err != nil {
			return nil, err
		}
	}

	// 使用syscall的mmap接口，创建内存映射
	// mmap接口相比posix接口，少了一个addr参数，如果有需要可以使用syscall.Syscall6接口
	// MAP_SHARED指定映射的类型，该模式下对映射空间的更新对其他进程的映射可见，并且会写回底层文件
	// 映射内存会通过[]byte的形式返回
	buf, err := syscall.Mmap(int(fd.Fd()), 0, int(size*2+headSize), syscall.PROT_WRITE|syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		return nil, err
	}

	// mmap返回之后，底层文件的设备描述符可以立即close掉
	fd.Close() // After the mmap() call has returned, the file descriptor can be closed immediately

	currentBlockIndex := (*int32)(unsafe.Pointer(&buf[0]))
	blockSize := (*uint64)(unsafe.Pointer(&buf[4]))
	tag1 := (*tag)(unsafe.Pointer(&buf[12]))
	tag2 := (*tag)(unsafe.Pointer(&buf[20]))

	// 与上一次数据的大小不一样
	if *blockSize != size {
		// judge the block
		var existData []byte
		base := uint64(headSize)
		tag := tag1
		if *currentBlockIndex == 1 {
			tag = tag2
			base += *blockSize
		}

		switch {
		case tag.ReadIndex == tag.WriteIndex:
		case tag.ReadIndex < tag.WriteIndex:
			existData = append(existData, buf[int32(base)+tag.ReadIndex:int32(base)+tag.WriteIndex]...)
		default:
			existData = append(existData, buf[int32(base)+tag.ReadIndex:*blockSize+base]...)
			existData = append(existData, buf[base:tag.WriteIndex+int32(base)]...)
		}

		if uint64(len(existData)) > size {
			return nil, errors.New("the size is unable to store old data")
		}

		oldShm := &SHM{
			currentBlockIndex: currentBlockIndex,
			size:              blockSize,
			mMaps: [2]*mMap{
				&mMap{
					tag:   tag1,
					queue: buf[headSize : headSize+int(*blockSize)],
				}, &mMap{
					tag:   tag2,
					queue: buf[headSize+int(*blockSize):],
				}},
		}

		// 如果是 block0
		if *oldShm.currentBlockIndex == 0 {
			if err := oldShm.Rewrite(existData); err != nil {
				return nil, err
			}
		}
		if err := oldShm.Rewrite(existData); err != nil {
			return nil, err
		}

		*blockSize = size
	}

	return &SHM{
		currentBlockIndex: currentBlockIndex,
		size:              blockSize,
		mMaps: [2]*mMap{
			&mMap{
				tag:   tag1,
				queue: buf[headSize : headSize+int(size)],
			}, &mMap{
				tag:   tag2,
				queue: buf[headSize+int(size):],
			}},
	}, nil
}

func (s *SHM) Save(buf []byte) error {
	return s.getBlock().save(buf)
}

func (s *SHM) Get() []byte {
	return s.getBlock().get()
}

func (s *SHM) Rewrite(buf []byte) error {
	unusedBlock := s.getUnusedBlock()
	unusedBlock.flush()
	if err := unusedBlock.save(buf); err != nil {
		return err
	}

	s.changeBlock()
	s.getUnusedBlock().flush()
	return nil
}

// getBlock return the block in use
func (s *SHM) getBlock() *mMap {
	if *s.currentBlockIndex == 0 {
		return s.mMaps[0]
	}
	return s.mMaps[1]
}

// getUnusedBlock return the unused block
func (s *SHM) getUnusedBlock() *mMap {
	if *s.currentBlockIndex == 0 {
		return s.mMaps[1]
	}
	return s.mMaps[0]
}

func (s *SHM) changeBlock() {
	if *s.currentBlockIndex == 0 {
		atomic.AddInt32(s.currentBlockIndex, 1)
	} else {
		atomic.AddInt32(s.currentBlockIndex, -1)
	}
}
