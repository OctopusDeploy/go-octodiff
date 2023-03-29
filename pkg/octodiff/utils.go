package octodiff

import (
	"encoding/binary"
	"fmt"
	"io"
)

// There are a number of places where we allocate buffers to read files, which use this value
const defaultReadBufferSize = 4 * 1024 * 1024

// returns the string, how many bytes we read in order to get it, and an error
func readLengthPrefixedString(input io.Reader) (string, int, error) {
	// C# BinaryWriter prefixes strings with their length using a single byte for small strings, or 4 bytes for larger
	// We only handle small strings here
	var contentLen uint8
	err := binary.Read(input, binary.LittleEndian, &contentLen)
	if err != nil {
		return "", 0, err
	}

	var content = make([]byte, contentLen)
	bytesRead, err := input.Read(content)
	if err != nil {
		return "", 1 + bytesRead, err
	}
	if bytesRead != int(contentLen) {
		return "", 1 + bytesRead, fmt.Errorf("Binary format indicates string length to read of %d but only %d bytes were read", contentLen, bytesRead)
	}
	return string(content), 1 + bytesRead, nil
}

func writeLengthPrefixedString(output io.Writer, str string) error {
	// C# BinaryWriter prefixes strings with their length using a single byte for small strings, or 4 bytes for larger
	// We only handle small strings here
	strBytes := []byte(str)
	_, err := output.Write([]byte{byte(len(strBytes))})
	if err != nil {
		return err
	}
	_, err = output.Write(strBytes)
	return err
}
