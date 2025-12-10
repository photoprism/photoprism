package fs

import "sync"

// copyBufferSize defines the shared buffer length used for large file copies/hashes.
const copyBufferSize = 256 * 1024

// copyBufferPool reuses byte slices to reduce allocations during file I/O.
var copyBufferPool = sync.Pool{ //nolint:gochecknoglobals // shared pool for I/O buffers
	New: func() any {
		buf := make([]byte, copyBufferSize)
		return &buf
	},
}

// getCopyBuffer returns a pooled buffer for copy operations.
func getCopyBuffer() []byte {
	return *(copyBufferPool.Get().(*[]byte))
}

// putCopyBuffer returns a buffer to the pool.
func putCopyBuffer(buf []byte) {
	copyBufferPool.Put(&buf)
}
