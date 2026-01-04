package exec

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"syscall"
	"time"
	"unsafe"
)

// ErrWinDivertMissing indicates WinDivert driver/runtime is not available.
var ErrWinDivertMissing = errors.New("WinDivert driver not available; install WinDivert to enable network injection")

// NetworkDelayRunner describes delay/loss/bandwidth shaping parameters.
type NetworkDelayRunner struct {
	DelayMillis   int
	JitterMillis  int
	LossPercent   float64
	BandwidthKbps int
	Filter        string
}

const defaultNetFilter = "outbound and tcp"

// NewNetworkDelayRunner creates a runner with given shaping parameters.
func NewNetworkDelayRunner(delayMillis, jitterMillis int, lossPercent float64, filter string, bandwidthKbps int) *NetworkDelayRunner {
	return &NetworkDelayRunner{
		DelayMillis:   delayMillis,
		JitterMillis:  jitterMillis,
		LossPercent:   lossPercent,
		BandwidthKbps: bandwidthKbps,
		Filter:        filter,
	}
}

// Run applies delay/loss/bandwidth shaping using WinDivert.
func (r *NetworkDelayRunner) Run(ctx context.Context) error {
	if r.DelayMillis < 0 || r.JitterMillis < 0 || r.LossPercent < 0 || r.LossPercent > 100 {
		return fmt.Errorf("invalid network params: delay=%d jitter=%d loss=%.2f", r.DelayMillis, r.JitterMillis, r.LossPercent)
	}

	if r.Filter == "" {
		r.Filter = defaultNetFilter
	}

	if err := loadWinDivert(); err != nil {
		return err
	}

	handle, err := winDivertOpen(r.Filter)
	if err != nil {
		if errors.Is(err, syscall.ERROR_FILE_NOT_FOUND) {
			return ErrWinDivertMissing
		}
		return err
	}
	defer winDivertClose(handle)

	// Ensure any blocking recv is released when context ends.
	go func() {
		<-ctx.Done()
		_ = winDivertShutdown(handle)
	}()

	pktBuf := make([]byte, 1<<16) // 64 KiB for packet payloads
	addrBuf := make([]byte, 128)  // WINDIVERT_ADDRESS opaque storage (use larger buffer to be safe)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	baseDelay := time.Duration(r.DelayMillis) * time.Millisecond
	jitter := time.Duration(r.JitterMillis) * time.Millisecond

	var rateBytesPerSec float64
	if r.BandwidthKbps > 0 {
		rateBytesPerSec = float64(r.BandwidthKbps) * 1000.0 / 8.0
	}

	timer := time.NewTimer(time.Hour)
	timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		n, addrLen, recvErr := winDivertRecv(handle, pktBuf, addrBuf)
		if recvErr != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return recvErr
		}

		if r.LossPercent > 0 && rng.Float64()*100.0 < r.LossPercent {
			continue
		}

		delay := baseDelay
		if jitter > 0 {
			// Uniform jitter in [-jitter, +jitter].
			offset := time.Duration(rng.Int63n(int64(jitter)*2)) - jitter
			delay += offset
			if delay < 0 {
				delay = 0
			}
		}

		if rateBytesPerSec > 0 {
			bwDelay := time.Duration(float64(n) * float64(time.Second) / rateBytesPerSec)
			delay += bwDelay
		}

		if delay > 0 {
			timer.Reset(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
		}

		if err := winDivertSend(handle, pktBuf[:n], addrBuf[:addrLen]); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return err
		}
	}
}

// --- Minimal WinDivert binding ---

const (
	winDivertLayerNetwork = 0
	winDivertShutdownBoth = 2
)

var (
	winDivertDLL          = syscall.NewLazyDLL("WinDivert.dll")
	procWinDivertOpen     = winDivertDLL.NewProc("WinDivertOpen")
	procWinDivertRecv     = winDivertDLL.NewProc("WinDivertRecv")
	procWinDivertSend     = winDivertDLL.NewProc("WinDivertSend")
	procWinDivertClose    = winDivertDLL.NewProc("WinDivertClose")
	procWinDivertShutdown = winDivertDLL.NewProc("WinDivertShutdown")
)

func loadWinDivert() error {
	if err := winDivertDLL.Load(); err != nil {
		if errors.Is(err, syscall.ERROR_MOD_NOT_FOUND) || errors.Is(err, syscall.Errno(193)) {
			// missing DLL or wrong arch
			return ErrWinDivertMissing
		}
		return err
	}
	return nil
}

func winDivertOpen(filter string) (syscall.Handle, error) {
	filterPtr, err := syscall.BytePtrFromString(filter)
	if err != nil {
		return 0, err
	}

	h, _, callErr := procWinDivertOpen.Call(
		uintptr(unsafe.Pointer(filterPtr)),
		uintptr(winDivertLayerNetwork),
		uintptr(int16(0)),
		uintptr(uint64(0)),
	)
	// WinDivertOpen returns NULL or INVALID_HANDLE_VALUE (-1) on failure
	if h == 0 || h == ^uintptr(0) {
		// Return a more detailed error including the underlying syscall error
		if callErr != nil {
			if errno, ok := callErr.(syscall.Errno); ok {
				return 0, fmt.Errorf("WinDivertOpen failed: %w (errno=%d)", errno, int(errno))
			}
			return 0, fmt.Errorf("WinDivertOpen failed: %v", callErr)
		}
		return 0, fmt.Errorf("WinDivertOpen failed: unknown error (handle invalid)")
	}
	// Debug: print handle value
	fmt.Printf("DEBUG: WinDivertOpen returned handle=0x%X\n", h)
	return syscall.Handle(h), nil
}

func winDivertRecv(h syscall.Handle, pktBuf []byte, addrBuf []byte) (int, int, error) {
	var recvLen uint64
	addrLen := uint32(len(addrBuf))

	// Debug: print buffer sizes before call
	fmt.Printf("DEBUG: WinDivertRecv called with handle=0x%X pktBufLen=%d addrBufLen=%d\n", h, len(pktBuf), len(addrBuf))

	r1, _, err := procWinDivertRecv.Call(
		uintptr(h),
		uintptr(unsafe.Pointer(&pktBuf[0])),
		uintptr(len(pktBuf)),
		uintptr(unsafe.Pointer(&recvLen)),
		uintptr(unsafe.Pointer(&addrBuf[0])),
		uintptr(unsafe.Pointer(&addrLen)),
		uintptr(uint64(0)),
	)
	if r1 == 0 {
		if err != nil {
			if errno, ok := err.(syscall.Errno); ok {
				return 0, 0, fmt.Errorf("WinDivertRecv failed: %w (errno=%d)", errno, int(errno))
			}
			return 0, 0, fmt.Errorf("WinDivertRecv failed: %v", err)
		}
		return 0, 0, fmt.Errorf("WinDivertRecv failed: unknown error (r1==0)")
	}
	return int(recvLen), int(addrLen), nil
}

func winDivertSend(h syscall.Handle, pkt []byte, addr []byte) error {
	var sendLen uint64
	addrLen := uint32(len(addr))

	r1, _, err := procWinDivertSend.Call(
		uintptr(h),
		uintptr(unsafe.Pointer(&pkt[0])),
		uintptr(len(pkt)),
		uintptr(unsafe.Pointer(&sendLen)),
		uintptr(unsafe.Pointer(&addr[0])),
		uintptr(unsafe.Pointer(&addrLen)),
		uintptr(uint64(0)),
	)
	if r1 == 0 {
		if err != nil {
			if errno, ok := err.(syscall.Errno); ok {
				return fmt.Errorf("WinDivertSend failed: %w (errno=%d)", errno, int(errno))
			}
			return fmt.Errorf("WinDivertSend failed: %v", err)
		}
		return fmt.Errorf("WinDivertSend failed: unknown error (r1==0)")
	}
	if int(sendLen) != len(pkt) {
		return fmt.Errorf("partial send: %d/%d", sendLen, len(pkt))
	}
	return nil
}

func winDivertShutdown(h syscall.Handle) error {
	r1, _, err := procWinDivertShutdown.Call(uintptr(h), uintptr(winDivertShutdownBoth))
	if r1 == 0 && err != nil {
		return err
	}
	return nil
}

func winDivertClose(h syscall.Handle) {
	procWinDivertClose.Call(uintptr(h))
}
