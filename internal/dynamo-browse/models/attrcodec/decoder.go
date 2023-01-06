package attrcodec

import (
	"encoding/binary"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/pkg/errors"
	"io"
)

type Decoder struct {
	r io.Reader
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

func (d *Decoder) Decode() (types.AttributeValue, error) {
	return d.decode()
}

func (d *Decoder) decode() (types.AttributeValue, error) {
	fr, err := d.readFrame()
	if err != nil {
		return nil, err
	}

	switch fr.typeID {
	case typeString:
		return &types.AttributeValueMemberS{Value: string(fr.data)}, nil
	case typeNumber:
		return &types.AttributeValueMemberN{Value: string(fr.data)}, nil
	case typeBoolean:
		return &types.AttributeValueMemberBOOL{Value: fr.flags&flagsAlternative != 0}, nil
	case typeNull:
		return &types.AttributeValueMemberNULL{Value: fr.flags&flagsAlternative == 0}, nil
	case typeList:
		vals := make([]types.AttributeValue, fr.length)
		for i := range vals {
			v, err := d.decode()
			if err != nil {
				return nil, err
			}
			vals[i] = v
		}
		return &types.AttributeValueMemberL{Value: vals}, nil
	case typeMap:
		vals := make(map[string]types.AttributeValue)
		for i := 0; i < fr.length; i++ {
			// key
			keyFrame, err := d.readFrame()
			if err != nil {
				return nil, err
			} else if keyFrame.typeID != typeString {
				return nil, errors.Errorf("key of %v must be string, but is ID %v", i, keyFrame.typeID)
			}

			// value
			v, err := d.decode()
			if err != nil {
				return nil, err
			}
			vals[string(keyFrame.data)] = v
		}
		return &types.AttributeValueMemberM{Value: vals}, nil
	}

	return nil, errors.Errorf("unrecognised type ID: %x", fr.typeID)
}

func (d *Decoder) readFrame() (frame, error) {
	var typeBfr [1]byte

	n, err := d.r.Read(typeBfr[:])
	if err != nil {
		return frame{}, err
	} else if n != 1 {
		return frame{}, errors.New("expected frame typeID")
	}

	typeID := typeBfr[0] &^ flagMask
	flags := typeBfr[0] & flagMask

	typeInfo, hasTypeInfo := typeFrameInfos[typeID]
	if !hasTypeInfo {
		return frame{}, errors.Errorf("unrecognised typeID: %x", typeID)
	}

	if typeInfo.isNilLength {
		return frame{typeID: typeID, flags: flags, data: nil}, nil
	}

	// TODO: this needs to depend on the type
	var l int64
	if flags&flagsAlternative != 0 {
		if err := binary.Read(d.r, byteOrder, &l); err != nil {
			return frame{}, errors.Wrap(err, "cannot encode alt length")
		}
	} else {
		var lenBfr [1]byte

		n, err := d.r.Read(lenBfr[:])
		if err != nil {
			return frame{}, err
		} else if n != 1 {
			return frame{}, errors.New("expected frame typeID")
		}
		l = int64(lenBfr[0])
	}

	if typeInfo.lengthOnly {
		return frame{typeID: typeID, flags: flags, length: int(l)}, nil
	}

	bs := make([]byte, l)
	n, err = d.r.Read(bs)
	if err != nil {
		return frame{}, err
	} else if n != int(l) {
		return frame{}, errors.Errorf("expected %v bytes but received %v", l, n)
	}

	return frame{typeID: typeID, flags: flags, data: bs}, nil
}
