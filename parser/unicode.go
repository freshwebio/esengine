package parser

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"unicode/utf16"
)

// DecodeUnicodeEscSeq deals with decoding
// a unicode escape sequence into the actual represented
// code points.
func DecodeUnicodeEscSeq(input []rune) []rune {
	var encoded []uint16
	var sequenceStrings []string
	var sequenceStr string
	j := 0
	for i := 0; i < len(input); i++ {
		if input[i] != '\\' && input[i] != 'u' && input[i] != '{' && input[i] != '}' {
			sequenceStr += string(input[i])
			if (j+1)%4 == 0 {
				sequenceStrings = append(sequenceStrings, sequenceStr)
				sequenceStr = ""
			}
			j++
		}
	}
	for i := 0; i < len(sequenceStrings); i++ {
		sequenceStr = sequenceStrings[i]
		hexBytes, err := hex.DecodeString(sequenceStr)
		if err != nil {
			panic(err)
		}
		encodedUint16 := binary.BigEndian.Uint16(hexBytes)
		fmt.Println(encodedUint16)
		encoded = append(encoded, encodedUint16)
	}
	decoded := utf16.Decode(encoded)
	return decoded
}
