package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

// IsElevated checks whether the current process is running with elevated privileges.
func IsElevated() bool {
	// Use OpenProcessToken + GetTokenInformation(TokenElevation)
	var h syscall.Handle
	procOpenProcessToken := syscall.NewLazyDLL("advapi32.dll").NewProc("OpenProcessToken")
	procGetTokenInformation := syscall.NewLazyDLL("advapi32.dll").NewProc("GetTokenInformation")

	// Current process pseudo-handle
	pid, _ := syscall.GetCurrentProcess()
	ret, _, _ := procOpenProcessToken.Call(uintptr(pid), uintptr(syscall.TOKEN_QUERY), uintptr(unsafe.Pointer(&h)))
	if ret == 0 {
		return false
	}
	defer syscall.CloseHandle(h)

	var elevation uint32
	var outLen uint32
	r, _, _ := procGetTokenInformation.Call(uintptr(h), uintptr(20), uintptr(unsafe.Pointer(&elevation)), uintptr(unsafe.Sizeof(elevation)), uintptr(unsafe.Pointer(&outLen)))
	if r == 0 {
		return false
	}
	return elevation != 0
}

// EnsureElevated will relaunch elevated if the current process is not elevated.
// It returns true if the current process should exit because a relaunch was attempted.
func EnsureElevated() bool {
	if IsElevated() {
		return false
	}
	fmt.Fprintln(os.Stderr, "This command requires administrator privileges. Requesting elevation...")
	if err := relaunchElevated(); err == nil {
		return true
	}
	fmt.Fprintln(os.Stderr, "Failed to relaunch elevated. Please run the command in an Administrator shell.")
	return false
}

// RequestElevationIfNeeded examines an error and, if it appears to be a permission
// denied error, prompts the user and attempts to relaunch the current executable
// with elevated (administrator) privileges. Returns true if a relaunch was
// attempted (caller should exit), false otherwise.
func RequestElevationIfNeeded(err error) bool {
	if err == nil {
		return false
	}
	// Quick string check for common message plus syscall check
	msg := err.Error()
	if strings.Contains(msg, "Access is denied") || strings.Contains(msg, "permission") || strings.Contains(msg, "access") || strings.Contains(msg, syscall.Errno(5).Error()) {
		fmt.Fprintln(os.Stderr, "Permission denied. Attempting to relaunch with administrator privileges...")
		if relaunchElevated() == nil {
			return true
		}
		fmt.Fprintln(os.Stderr, "Failed to relaunch elevated. Please run the command in an Administrator shell.")
	}
	return false
}

// relaunchElevated uses ShellExecute to request elevation for the current binary.
func relaunchElevated() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	// Build args from os.Args using strconv.Quote for safe quoting
	args := ""
	if len(os.Args) > 1 {
		for i, a := range os.Args[1:] {
			if i > 0 {
				args += " "
			}
			args += strconv.Quote(a)
		}
	}

	// Use ShellExecuteW via rundll32-ish approach: use 'runas' verb with ShellExecute
	verbPtr, _ := syscall.UTF16PtrFromString("runas")
	filePtr, _ := syscall.UTF16PtrFromString(exe)
	paramsPtr, _ := syscall.UTF16PtrFromString(args)
	// Pass current working directory to ShellExecuteW
	cwd, _ := os.Getwd()
	cwdPtr, _ := syscall.UTF16PtrFromString(cwd)
	// ShellExecuteW returns an HINST; use syscall to call it
	shell32 := syscall.NewLazyDLL("shell32.dll")
	proc := shell32.NewProc("ShellExecuteW")
	r, _, e := proc.Call(
		0,
		uintptr(unsafe.Pointer(verbPtr)),
		uintptr(unsafe.Pointer(filePtr)),
		uintptr(unsafe.Pointer(paramsPtr)),
		uintptr(unsafe.Pointer(cwdPtr)),
		uintptr(1), // SW_SHOWNORMAL
	)
	if r <= 32 {
		if e != nil {
			return e
		}
		return fmt.Errorf("ShellExecuteW failed: return=%d", r)
	}
	return nil
}
