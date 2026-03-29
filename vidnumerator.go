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

	// V4L2CapVideoCapture is the device capability flag indicating
	// the device supports video capture (V4L2_CAP_VIDEO_CAPTURE | V4L2_CAP_STREAMING | V4L2_CAP_DEVICE_CAPS).
	V4L2CapVideoCapture uint32 = 69206017
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

// this function checks the ioctl for VIDIOC_QUERYCAP to see if the device is a video capture device
func IsVideoCapture(path string) (bool, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0o755)
	if err != nil {
		return false, err
	}
	fd := f.Fd()
	ic := cap{}
	err = ic.QueryFd(int(fd))
	if err != nil {
		return false, err
	}
	return ic.deviceCaps == V4L2CapVideoCapture, nil
}

// this function checks the ioctl for VIDIOC_QUERYCAP to see if the device is a video capture device
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
