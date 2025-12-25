WARNING: Partial implementation, don't use it in production. If you are searching for a good implementation,
use [go-rtmp](https://github.com/yutopp/go-rtmp) as it is confirmed to be the one used by Twitch

A RTMP server protocol implementation according to the specifications:

[RTMP specification](https://veovera.org/docs/legacy/rtmp-v1-0-spec.pdf)

[AMF0 specification](https://rtmp.veriskope.com/pdf/amf0-file-format-specification.pdf)

Missing features:

- Authentication, Encryption
- Multiplexing
- Event handling
- Performance (OBS is reporting that the ack is too slow)