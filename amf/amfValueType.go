package amf

type AmfValueType interface {
	Encode() []byte
}

func decodeNextAmfType(bytes []byte) (int, AmfValueType) {
	valueTypeMarker := bytes[0]
	var valueType AmfValueType
	var length int
	switch valueTypeMarker {
	case numberMarker:
		length, valueType = decodeNextAmfNumber(bytes)
	case stringMarker:
		length, valueType = decodeNextAmfString(bytes)
	case booleanMarker:
		length, valueType = decodeNextAmfBoolean(bytes)
	case objectMarker:
		length, valueType, _ = decodeNextAmfObject(bytes)
	default:
		return 0, nil
	}
	return length, valueType
}
