package exec

import (
	"context"
	"time"
)

// MemoryRunner allocates and holds memory until the context is canceled.
type MemoryRunner struct {
	sizeBytes int64
}

// NewMemoryRunner builds a MemoryRunner for the requested size in bytes.
func NewMemoryRunner(sizeBytes int64) *MemoryRunner {
	const minBytes = 1 << 20 // 1 MB safeguard
	if sizeBytes < minBytes {
		sizeBytes = minBytes
	}
	return &MemoryRunner{sizeBytes: sizeBytes}
}

// Run allocates, touches pages, and keeps the memory until cancellation.
func (r *MemoryRunner) Run(ctx context.Context) error {
	buf := make([]byte, r.sizeBytes)

	for i := int64(0); i < r.sizeBytes; i += 4096 {
		buf[i] = byte(i)
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if len(buf) > 0 {
				buf[0] ^= 1
			}
		}
	}
}
