package vidnumerator

import (
	"os"
	"testing"
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

func TestCapIsVideoCaptureAcceptsAdditionalFlags(t *testing.T) {
	ic := cap{
		capabilities: V4L2CapDeviceCaps,
		deviceCaps: V4L2CapVideoCapture |
			V4L2CapStreaming |
			0x00000002,
	}

	if !ic.isVideoCapture() {
		t.Fatal("expected device with capture and streaming flags to be detected")
	}
}

func TestCapIsVideoCaptureRequiresCaptureAndStreaming(t *testing.T) {
	tests := []struct {
		name string
		ic   cap
	}{
		{
			name: "missing capture",
			ic: cap{
				capabilities: V4L2CapDeviceCaps,
				deviceCaps:   V4L2CapStreaming,
			},
		},
		{
			name: "missing streaming",
			ic: cap{
				capabilities: V4L2CapDeviceCaps,
				deviceCaps:   V4L2CapVideoCapture,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.ic.isVideoCapture() {
				t.Fatal("expected non-capture device")
			}
		})
	}
}
