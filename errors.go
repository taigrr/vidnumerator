package vidnumerator

import "errors"

// ErrUnsupportedPlatform indicates that V4L2 video device enumeration is only available on Linux.
var ErrUnsupportedPlatform = errors.New("vidnumerator: unsupported platform")
