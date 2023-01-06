package attrcodec

import (
	"encoding/binary"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/pkg/errors"
	"io"
)

var byteOrder = binary.LittleEndian

type Encoder struct {
	w io.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

func (e *Encoder) Encode(val types.AttributeValue) error {
	return e.encode(val)
}

func (e *Encoder) encode(val types.AttributeValue) error {
	switch v := val.(type) {
	case *types.AttributeValueMemberS:
		return e.writeFrame(typeString, []byte(v.Value))
	case *types.AttributeValueMemberN:
		return e.writeFrame(typeNumber, []byte(v.Value))
	case *types.AttributeValueMemberBOOL:
		if v.Value {
			return e.writeNilLengthFrame(typeBoolean, flagsAlternative)
		} else {
			return e.writeNilLengthFrame(typeBoolean, 0x0)
		}
	case *types.AttributeValueMemberNULL:
		if !v.Value {
			return e.writeNilLengthFrame(typeNull, flagsAlternative)
		} else {
			return e.writeNilLengthFrame(typeNull, 0x0)
		}
	case *types.AttributeValueMemberL:
		if err := e.writeFrameHeader(typeList, len(v.Value)); err != nil {
			return err
		}
		for _, nv := range v.Value {
			if err := e.encode(nv); err != nil {
				return err
			}
		}
		return nil
	case *types.AttributeValueMemberM:
		if err := e.writeFrameHeader(typeMap, len(v.Value)); err != nil {
			return err
		}

		for k, kv := range v.Value {
			// Keys are always strings
			if err := e.writeFrame(typeString, []byte(k)); err != nil {
				return err
			}
			if err := e.encode(kv); err != nil {
				return err
			}
		}
		return nil
	}
	return errors.New("unhandled type")
}

func (e *Encoder) writeNilLengthFrame(typeID byte, flags byte) error {
	if _, err := e.w.Write([]byte{typeID | flags}); err != nil {
		return err
	}
	return nil
}

func (e *Encoder) writeFrameHeader(typeID byte, length int) error {
	if length <= 255 {
		if _, err := e.w.Write([]byte{typeID, byte(length)}); err != nil {
			return err
		}

		return nil
	}

	// Length longer than a byte, use a int32
	if _, err := e.w.Write([]byte{typeID | flagsAlternative}); err != nil {
		return err
	}

	if err := binary.Write(e.w, byteOrder, int64(length)); err != nil {
		return errors.Wrap(err, "cannot encode alt length")
	}

	return nil
}

func (e *Encoder) writeFrame(typeID byte, bts []byte) error {
	if err := e.writeFrameHeader(typeID, len(bts)); err != nil {
		return err
	}

	_, err := e.w.Write(bts)
	return err
}
