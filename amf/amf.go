package amf

import (
	"errors"
	"iter"
)

type AmfCommand struct {
	parts []AmfValueType
}

func (command AmfCommand) Encode() []byte {
	bytes := make([]byte, 0)
	for _, part := range command.parts {
		bytes = append(bytes, part.Encode()...)
	}
	return bytes
}

func NewAmfCommand(parts ...AmfValueType) AmfCommand {
	return AmfCommand{
		parts: parts,
	}
}

func DecodeAmfCommand(bytes []byte) (*AmfCommand, error) {
	command := AmfCommand{
		parts: make([]AmfValueType, 0),
	}
	for part := range decodeParts(bytes) {
		if part == nil {
			return nil, errors.New("can't decode command parts, invalid marker")
		}
		command.parts = append(command.parts, part)
	}
	return &command, nil
}

func decodeParts(bytes []byte) iter.Seq[AmfValueType] {

	return func(yield func(AmfValueType) bool) {
		pointer := 0
		for pointer < len(bytes) {
			valueLength, valueType := decodeNextAmfType(bytes[pointer:])
			pointer += valueLength
			if !yield(valueType) {
				return
			}
		}

	}
}
