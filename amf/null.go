package amf

var nullMarker = byte(0x05)

type Null struct{}

func (null Null) Encode() []byte {
	return []byte{nullMarker}
}

func NewNull() Null {
	return Null{}
}

func decodeNextNull() (int, Null) {
	return 1, Null{}
}
