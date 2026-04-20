# VidNumerator

This is a tiny library that uses syscalls to efficiently determine which `/dev/videoN` devices
are webcams and which are the additional metadata control handles.
The list of strings returned are the full filepaths to valid devices.

>[!IMPORTANT]
>In order for this library to work properly, the executing user must have either root or video group privileges

## Technical Details

The library works by:

1. Scanning `/dev` for files matching the pattern `video*`
2. Calling the `VIDIOC_QUERYCAP` ioctl for each candidate
3. Reading the effective capability bits from `deviceCaps` when `V4L2_CAP_DEVICE_CAPS` is set, otherwise falling back to `capabilities`
4. Keeping only devices that advertise both `V4L2_CAP_VIDEO_CAPTURE` and `V4L2_CAP_STREAMING`

That filtering excludes metadata-only handles and other non-capture nodes that still show up as `/dev/videoN`.

## Usage

```go
devices, err := vidnumerator.EnumeratedVideoDevices()
if err != nil {
    // likely permission or ioctl failure
    log.Fatal(err)
}

for _, device := range devices {
    fmt.Println(device)
}
```

## Implementation Notes

- Uses direct syscalls via `golang.org/x/sys/unix`
- Implements custom ioctl constants for V4L2 device querying
- Uses capability bit checks instead of matching one exact integer value, so drivers with additional flags still work correctly
