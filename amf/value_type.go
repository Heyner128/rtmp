package amf

type ValueType interface {
	Encode() []byte
}

func decodeNextValueType(bytes []byte) (int, ValueType) {
	valueTypeMarker := bytes[0]
	var valueType ValueType
	var length int
	switch valueTypeMarker {
	case numberMarker:
		length, valueType = decodeNextNumber(bytes)
	case stringMarker:
		length, valueType = decodeNextString(bytes)
	case booleanMarker:
		length, valueType = decodeNextBoolean(bytes)
	case objectMarker:
		length, valueType, _ = decodeNextObject(bytes)
	default:
		return 0, nil
	}
	return length, valueType
}
