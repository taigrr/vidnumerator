# VidNumerator

This is a tiny library that uses syscalls to efficiently determine which `/dev/videoN` devices
are webcams and which are the additional metadata control handles.
The list of strings returned are the full filepaths to valid devices.

