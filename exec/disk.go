package exec

import (
	"context"
	"errors"
	"os"
)

// DiskFillRunner writes data to a file until a target size, then holds it until cancellation.
type DiskFillRunner struct {
	path      string
	sizeBytes int64
}

// NewDiskFillRunner builds a DiskFillRunner for a specific byte size.
func NewDiskFillRunner(path string, sizeBytes int64) *DiskFillRunner {
	if sizeBytes < 1 {
		sizeBytes = 1
	}
	return &DiskFillRunner{path: path, sizeBytes: sizeBytes}
}

// Run fills the target file and removes it when the context is canceled.
func (r *DiskFillRunner) Run(ctx context.Context) error {
	targetPath := r.path
	if targetPath == "" {
		f, err := os.CreateTemp("", "chaosblade-disk-*.tmp")
		if err != nil {
			return err
		}
		targetPath = f.Name()
		f.Close()
	}

	f, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer func() {
		f.Close()
		os.Remove(targetPath)
	}()

	buf := make([]byte, 1024*1024)
	var written int64

	for written < r.sizeBytes {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		remaining := r.sizeBytes - written
		chunk := len(buf)
		if int64(chunk) > remaining {
			chunk = int(remaining)
		}

		n, err := f.Write(buf[:chunk])
		if err != nil {
			return err
		}
		written += int64(n)
	}

	if err := f.Sync(); err != nil {
		return err
	}

	<-ctx.Done()
	return ctx.Err()
}

// ErrDiskPathRequired indicates a missing path when required.
var ErrDiskPathRequired = errors.New("disk path required")

// DiskIOPSRunner placeholder for future IO stress.
type DiskIOPSRunner struct{}

func (DiskIOPSRunner) Run(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}
