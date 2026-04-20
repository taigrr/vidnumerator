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

	V4L2CapVideoCapture uint32 = 0x00000001
	V4L2CapStreaming    uint32 = 0x04000000
	V4L2CapDeviceCaps   uint32 = 0x80000000
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

func (r *cap) QueryFd(fileDescriptor int) error {
	if r == nil {
		return fmt.Errorf("nil receiver")
	}
	_, _, errorNumber := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(fileDescriptor),
		VidIOCQueryCap,
		uintptr(unsafe.Pointer(r)),
	)
	if errorNumber != 0 {
		return errorNumber
	}
	return nil
}

func (r cap) videoCaptureCaps() uint32 {
	if r.capabilities&V4L2CapDeviceCaps != 0 {
		return r.deviceCaps
	}
	return r.capabilities
}

func (r cap) isVideoCapture() bool {
	caps := r.videoCaptureCaps()
	return caps&V4L2CapVideoCapture != 0 && caps&V4L2CapStreaming != 0
}

// IsVideoCapture checks the ioctl for VIDIOC_QUERYCAP to see if the device is a video capture device.
func IsVideoCapture(path string) (bool, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return false, err
	}
	defer func() {
		_ = f.Close()
	}()

	ic := cap{}
	err = ic.QueryFd(int(f.Fd()))
	if err != nil {
		return false, err
	}
	return ic.isVideoCapture(), nil
}

// EnumeratedVideoDevices lists all /dev/video* nodes that support video capture.
func EnumeratedVideoDevices() ([]string, error) {
	// list all files in the /dev directory
	d, err := os.ReadDir("/dev")
	if err != nil {
		return []string{}, err
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
		isVidCap, err := IsVideoCapture(fname)
		if err != nil {
			return []string{}, err
		}
		if isVidCap {
			devNames = append(devNames, fname)
		}
	}
	return devNames, nil
}
