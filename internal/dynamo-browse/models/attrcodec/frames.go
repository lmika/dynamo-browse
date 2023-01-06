package attrcodec

const (
	typeString  byte = 0x01
	typeNumber  byte = 0x02
	typeBoolean byte = 0x03
	typeNull    byte = 0x04
	typeList    byte = 0x05
	typeMap     byte = 0x06

	flagMask = 0x80

	flagsAlternative = 0x80
)

type frame struct {
	typeID byte
	flags  byte
	length int
	data   []byte
}

type typeFrameInfo struct {
	isNilLength bool
	lengthOnly  bool
}

var typeFrameInfos = map[byte]typeFrameInfo{
	typeString:  {},
	typeNumber:  {},
	typeBoolean: {isNilLength: true},
	typeNull:    {isNilLength: true},
	typeList:    {lengthOnly: true},
	typeMap:     {lengthOnly: true},
}
