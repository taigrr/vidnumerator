package vidnumerator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"golang.org/x/sys/unix"
)

const (
	IOCNrBits   = 8
	IOCTypeBits = 8
	IOCSizeBits = 14
	IOCDirBits  = 2
	IOCNone     = 0
	IOCWrite    = 1
	IOCRead     = 2
	IOCNrShift  = 0

	IOCTypeShift = (IOCNrShift + IOCNrBits)
	IOCSizeShift = (IOCTypeShift + IOCTypeBits)
	IOCDirShift  = (IOCSizeShift + IOCSizeBits)

	VidIOCQueryCap = (IOCRead << IOCDirShift) |
		(uintptr('V') << IOCTypeShift) |
		(0 << IOCNrShift) |
		(unsafe.Sizeof(cap{}) << IOCSizeShift)
)

type cap struct {
	driver       [16]uint8
	card         [32]uint8
	busInfo      [32]uint8
	version      uint32
	capabilities uint32
	deviceCaps   uint32
	reserved     [3]uint32
}

func (r *cap) QueryFd(fileDesciptor int) error {
	if r == nil {
		return fmt.Errorf("nil receiver")
	}
	_, _, errorNumber := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(fileDesciptor),
		VidIOCQueryCap,
		uintptr(unsafe.Pointer(r)),
	)
	if errorNumber != 0 {
		return errorNumber
	}
	return nil
}

// this function checks the ioctl for VIDIOC_QUERYCAP to see if the device is a video capture device
func EnumeratedVideoDevices() []string {
	// list all files in the /dev directory
	d, err := os.ReadDir("/dev")
	if err != nil {
		return []string{}
	}
	// iterate over the files in the directory
	devNames := []string{}
	for _, file := range d {
		if file.IsDir() {
			continue
		}
		fname := file.Name()
		if !strings.HasPrefix(fname, "video") {
			continue
		}
		fname = filepath.Join("/dev/", fname)
		f, err := os.OpenFile(fname, os.O_RDONLY, 0o755)
		if err != nil {
			continue
		}
		fd := f.Fd()
		ic := cap{}
		err = ic.QueryFd(int(fd))
		if err != nil {
			continue
		}
		if ic.deviceCaps == 69206017 {
			devNames = append(devNames, fname)
		}
	}
	return devNames
}
