package vidnumerator

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"golang.org/x/sys/unix"
)

func TestCapQueryFdNilReceiver(t *testing.T) {
	var nilCap *cap
	err := nilCap.QueryFd(0)
	if err == nil {
		t.Fatal("expected error for nil receiver, got nil")
	}
	if err.Error() != "nil receiver" {
		t.Fatalf("expected 'nil receiver' error, got: %s", err)
	}
}

func TestCapQueryFdInvalidFd(t *testing.T) {
	ic := cap{}
	err := ic.QueryFd(-1)
	if err == nil {
		t.Fatal("expected error for invalid file descriptor, got nil")
	}
}

func TestIsVideoCaptureInvalidPath(t *testing.T) {
	isVid, err := IsVideoCapture("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for nonexistent path, got nil")
	}
	if isVid {
		t.Fatal("expected false for nonexistent path")
	}
}

func TestIsVideoCaptureRegularFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "vidnum-test-*")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	isVid, err := IsVideoCapture(tmpFile.Name())
	if err == nil {
		// Some kernels may return an error, some may not — both are valid
		if isVid {
			t.Fatal("regular file should not be detected as video capture")
		}
	}
}

func TestEnumeratedVideoDevices(t *testing.T) {
	// This test verifies the function runs without panic.
	// On machines without video devices, it should return an empty list.
	devices, err := EnumeratedVideoDevices()
	if err != nil {
		// /dev might not be readable in some CI environments
		t.Skipf("EnumeratedVideoDevices returned error (expected in some environments): %v", err)
	}
	// Just verify all returned paths start with /dev/video
	for _, device := range devices {
		if len(device) < 10 || device[:10] != "/dev/video" {
			t.Errorf("unexpected device path: %s", device)
		}
	}
}

func TestCapVideoCaptureCapsUsesDeviceCapsWhenPresent(t *testing.T) {
	ic := cap{
		capabilities: V4L2CapDeviceCaps,
		deviceCaps:   V4L2CapVideoCapture | V4L2CapStreaming,
	}

	if got := ic.videoCaptureCaps(); got != ic.deviceCaps {
		t.Fatalf("videoCaptureCaps() = %#x, want %#x", got, ic.deviceCaps)
	}
}

func TestCapVideoCaptureCapsFallsBackToCapabilities(t *testing.T) {
	ic := cap{
		capabilities: V4L2CapVideoCapture | V4L2CapStreaming,
	}

	if got := ic.videoCaptureCaps(); got != ic.capabilities {
		t.Fatalf("videoCaptureCaps() = %#x, want %#x", got, ic.capabilities)
	}
}

func TestCapIsVideoCapture(t *testing.T) {
	tests := []struct {
		name string
		caps uint32
		want bool
	}{
		{
			name: "single-planar capture with streaming",
			caps: V4L2CapVideoCapture | V4L2CapStreaming,
			want: true,
		},
		{
			name: "multi-planar capture with streaming",
			caps: V4L2CapVideoCaptureMPlane | V4L2CapStreaming,
			want: true,
		},
		{
			name: "capture with additional flags",
			caps: V4L2CapVideoCapture | V4L2CapStreaming | 0x00000002,
			want: true,
		},
		{
			name: "missing streaming",
			caps: V4L2CapVideoCapture,
			want: false,
		},
		{
			name: "missing capture",
			caps: V4L2CapStreaming,
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ic := cap{
				capabilities: V4L2CapDeviceCaps,
				deviceCaps:   test.caps,
			}

			if got := ic.isVideoCapture(); got != test.want {
				t.Fatalf("isVideoCapture() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestShouldSkipDeviceError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "enotty", err: unix.ENOTTY, want: true},
		{name: "einval", err: unix.EINVAL, want: true},
		{name: "enodev", err: unix.ENODEV, want: true},
		{name: "enoent", err: unix.ENOENT, want: true},
		{name: "wrapped enotty", err: fmt.Errorf("wrapped: %w", unix.ENOTTY), want: true},
		{name: "permission", err: os.ErrPermission, want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := shouldSkipDeviceError(test.err); got != test.want {
				t.Fatalf("shouldSkipDeviceError(%v) = %v, want %v", test.err, got, test.want)
			}
		})
	}
}

func TestEnumeratedVideoDevicesFromEntriesSkipsExpectedDeviceErrors(t *testing.T) {
	entries := []os.DirEntry{
		fakeDirEntry{name: "video0"},
		fakeDirEntry{name: "video1"},
		fakeDirEntry{name: "video2"},
		fakeDirEntry{name: "not-video"},
		fakeDirEntry{name: "video-dir", dir: true},
	}

	devices, err := enumeratedVideoDevicesFromEntries("/dev", entries, func(path string) (bool, error) {
		switch path {
		case "/dev/video0":
			return true, nil
		case "/dev/video1":
			return false, unix.ENOTTY
		case "/dev/video2":
			return false, nil
		default:
			return false, errors.New("unexpected path")
		}
	})
	if err != nil {
		t.Fatalf("enumeratedVideoDevicesFromEntries() error = %v", err)
	}
	if len(devices) != 1 || devices[0] != "/dev/video0" {
		t.Fatalf("enumeratedVideoDevicesFromEntries() = %v, want [/dev/video0]", devices)
	}
}

func TestEnumeratedVideoDevicesFromEntriesReturnsUnexpectedErrors(t *testing.T) {
	entries := []os.DirEntry{fakeDirEntry{name: "video0"}}
	expectedErr := errors.New("boom")

	_, err := enumeratedVideoDevicesFromEntries("/dev", entries, func(path string) (bool, error) {
		return false, expectedErr
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("enumeratedVideoDevicesFromEntries() error = %v, want %v", err, expectedErr)
	}
}

type fakeDirEntry struct {
	name string
	dir  bool
}

func (entry fakeDirEntry) Name() string               { return entry.name }
func (entry fakeDirEntry) IsDir() bool                { return entry.dir }
func (entry fakeDirEntry) Type() os.FileMode          { return 0 }
func (entry fakeDirEntry) Info() (os.FileInfo, error) { return nil, nil }
