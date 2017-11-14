package quicktemplate

import (
	"sync"
)

// ByteBuffer implements io.Writer on top of byte slice.
//
// Recycle byte buffers via AcquireByteBuffer and ReleaseByteBuffer
// in order to reduce memory allocations.
type ByteBuffer struct {
	// B is a byte slice backing byte buffer.
	// All the data written via Write is appended here.
	B []byte
}

// Write implements io.Writer.
func (bb *ByteBuffer) Write(p []byte) (int, error) {
	bb.B = append(bb.B, p...)
	return len(p), nil
}

// Reset resets the byte buffer.
func (bb *ByteBuffer) Reset() {
	bb.B = bb.B[:0]
}

// AcquireByteBuffer returns new ByteBuffer from the pool.
//
// Return unneeded buffers to the pool by calling ReleaseByteBuffer
// in order to reduce memory allocations.
func AcquireByteBuffer() *ByteBuffer {
	v := byteBufferPool.Get()
	if v == nil {
		return &ByteBuffer{}
	}
	return v.(*ByteBuffer)
}

// ReleaseByteBuffer retruns byte buffer to the pool.
//
// Do not access byte buffer after returning it to the pool,
// otherwise data races may occur.
func ReleaseByteBuffer(bb *ByteBuffer) {
	bb.Reset()
	byteBufferPool.Put(bb)
}

var byteBufferPool sync.Pool
