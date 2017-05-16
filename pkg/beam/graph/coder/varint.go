package coder

import (
	"errors"
	"io"
)

// Variable-length encoding for integers.
//
// Takes between 1 and 10 bytes. Less efficient for negative or large numbers.
// All negative ints are encoded using 5 bytes, longs take 10 bytes. We use
// uint64 (over int64) as the primitive form to get logical bit shifts.

var ErrVarIntTooLong = errors.New("varint too long")

// EncodeVarUint64 encodes an uint64.
func EncodeVarUint64(value uint64, w io.Writer) error {
	var ret []byte
	for {
		// Encode next 7 bits + terminator bit
		bits := value & 0x7f
		value >>= 7

		var mask uint64
		if value != 0 {
			mask = 0x80
		}
		ret = append(ret, (byte)(bits|mask))
		if value == 0 {
			_, err := w.Write(ret)
			return err
		}
	}
}

// TODO(herohde) 5/16/2017: figure out whether it's too slow to read one byte
// at a time here. If not, we may need a more sophisticated reader than
// io.Reader with lookahead, say.

// DecodeVarUint64 decodes an uint64.
func DecodeVarUint64(r io.Reader) (uint64, error) {
	var ret uint64
	var shift uint

	data := make([]byte, 1)
	for {
		// Get 7 bits from next byte
		if n, err := r.Read(data); n < 1 {
			return 0, err
		}

		b := data[0]
		bits := (uint64)(b & 0x7f)

		if shift >= 64 || (shift == 63 && bits > 1) {
			return 0, ErrVarIntTooLong
		}

		ret |= bits << shift
		shift += 7

		if (b & 0x80) == 0 {
			return ret, nil
		}
	}
}

// EncodeVarInt encodes an int32.
func EncodeVarInt(value int32, w io.Writer) error {
	return EncodeVarUint64((uint64)(value)&0xffffffff, w)
}

// DecodeVarInt decodes an int32.
func DecodeVarInt(r io.Reader) (int32, error) {
	ret, err := DecodeVarUint64(r)
	if err != nil {
		return 0, err
	}
	return (int32)(ret), nil
}
