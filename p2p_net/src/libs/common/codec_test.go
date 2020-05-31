package common

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
)

type test2 struct {
	data []byte
}

func (t test2) Marshal() []byte {
	return t.data
}

func unmarshalTest2(data []byte) (test2, error) {
	return test2{data: data}, nil
}

type test struct {
	data []byte
}

func (t test) Marshal() []byte {
	return t.data
}

func unmarshalTest(data []byte) (test, error) {
	return test{data: data}, nil
}

func TestCodecRegisterEncodeDecode(t *testing.T) {
	t.Parallel()

	codec := NewCodec()

	opCode := codec.Register(test{}, unmarshalTest)

	msg := test{data: []byte("hello world")}

	expected := make([]byte, 2+len(msg.data))
	binary.BigEndian.PutUint16(expected[:2], opCode)
	copy(expected[2:], msg.data)

	data, err := codec.Encode(msg)
	assert.NoError(t, err)

	assert.EqualValues(t, opCode, binary.BigEndian.Uint16(data[:2]))
	assert.EqualValues(t, expected, data)

	obj, err := codec.Decode(data)
	assert.NoError(t, err)
	assert.IsType(t, obj, test{})

	// Failure cases.

	data[0] = 99
	_, err = codec.Decode(data)
	assert.Error(t, err)

	_, err = codec.Encode(test2{data: []byte("should not be encoded")})
	assert.Error(t, err)

}

func TestPanicIfDuplicateMessagesRegistered(t *testing.T) {
	t.Parallel()

	codec := NewCodec()

	assert.Panics(t, func() {
		codec.Register(test{}, unmarshalTest)
		codec.Register(test2{}, unmarshalTest2)
		codec.Register(test{}, unmarshalTest)
	})
}
