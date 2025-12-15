package amf

import (
	"errors"
	"iter"
)

type Command struct {
	Parts []ValueType
}

func (command Command) Encode() []byte {
	bytes := make([]byte, 0)
	for _, part := range command.Parts {
		bytes = append(bytes, part.Encode()...)
	}
	return bytes
}

func NewCommand(parts ...ValueType) Command {
	return Command{
		Parts: parts,
	}
}

func DecodeCommand(bytes []byte) (*Command, error) {
	command := Command{
		Parts: make([]ValueType, 0),
	}
	for part := range decodeParts(bytes) {
		if part == nil {
			return nil, errors.New("can't decode command Parts, invalid marker")
		}
		command.Parts = append(command.Parts, part)
	}
	return &command, nil
}

func decodeParts(bytes []byte) iter.Seq[ValueType] {

	return func(yield func(ValueType) bool) {
		pointer := 0
		for pointer < len(bytes) {
			valueLength, valueType := decodeNextValueType(bytes[pointer:])
			pointer += valueLength
			if !yield(valueType) {
				return
			}
		}

	}
}
