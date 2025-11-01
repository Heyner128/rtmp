package amf

type AmfBoolean uint8

func newAmfBoolean(bool uint8) AmfBoolean {
	return AmfBoolean(bool)
}

func (bool AmfBoolean) encode() []byte {
	return []byte{uint8(bool)}
}

func decodeAmfBoolean(bytes []byte) AmfBoolean {
	return AmfBoolean(bytes[0])
}
