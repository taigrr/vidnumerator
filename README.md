# VidNumerator

This is a tiny library that uses syscalls to efficiently determine which `/dev/videoN` devices
are webcams and which are the additional metadata control handles.
The list of strings returned are the full filepaths to valid devices.

>[!IMPORTANT]
>In order for this library to work properly, the executing user must have either root or video group privileges

## Technical Details

The library works by:

1. Scanning `/dev` for files matching the pattern `video*`
2. Using the `VIDIOC_QUERYCAP` ioctl to check if each device is a video capture device
3. Filtering out non-capture devices (like metadata control handles)

The core functionality is implemented through direct syscalls to the Linux kernel's V4L2 (Video4Linux2) API. The library uses the `VIDIOC_QUERYCAP` ioctl command to query device capabilities and determine if a device supports video capture.

## Usage

```go
devices, err := vidnumerator.EnumeratedVideoDevices()
if err != nil {
    // handle error
}
// devices will contain paths like "/dev/video0", "/dev/video2", etc.
```

## Implementation Notes

- Uses direct syscalls via `golang.org/x/sys/unix`
- Implements custom ioctl constants for V4L2 device querying
- Checks for specific device capabilities (0x4200001) to identify capture devices
