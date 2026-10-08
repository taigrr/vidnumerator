//go:build !linux

package vidnumerator

// IsVideoCapture reports whether the path is a Linux V4L2 video capture device.
func IsVideoCapture(path string) (bool, error) {
	return false, ErrUnsupportedPlatform
}

// EnumeratedVideoDevices lists Linux V4L2 video capture devices.
func EnumeratedVideoDevices() ([]string, error) {
	return []string{}, ErrUnsupportedPlatform
}
