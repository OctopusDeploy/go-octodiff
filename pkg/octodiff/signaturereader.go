package octodiff

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

type SignatureReader struct {
	ProgressReporter ProgressReporter // must be non-null
}

func NewSignatureReader() *SignatureReader {
	return &SignatureReader{
		ProgressReporter: NopProgressReporter(),
	}
}

func (s *SignatureReader) ReadSignature(input io.Reader, inputLength int64) (*Signature, error) {
	pos := int64(0)
	s.ProgressReporter.ReportProgress("Reading signature", pos, inputLength)

	headerBytes := make([]byte, len(BinarySignatureHeader))
	bytesRead, err := input.Read(headerBytes)
	if err != nil {
		return nil, err
	}
	if bytesRead != len(BinarySignatureHeader) || !bytes.Equal(headerBytes, BinarySignatureHeader) {
		return nil, errors.New("the signature file appears to be corrupt")
	}
	pos += int64(bytesRead)

	var versionBytes = make([]byte, len(BinaryVersion))
	bytesRead, err = input.Read(versionBytes)
	if err != nil {
		return nil, err
	}
	if bytesRead != len(BinaryVersion) || !bytes.Equal(versionBytes, BinaryVersion) {
		return nil, errors.New("the signature file uses a newer file format than this program can handle")
	}
	pos += int64(bytesRead)

	hashAlgorithmStr, bytesRead, err := readLengthPrefixedString(input)
	if err != nil {
		return nil, err
	}
	pos += int64(bytesRead)

	rollingChecksumAlgorithmStr, bytesRead, err := readLengthPrefixedString(input)
	if err != nil {
		return nil, err
	}
	pos += int64(bytesRead)

	var endBytes = make([]byte, len(BinaryEndOfMetadata))
	bytesRead, err = input.Read(endBytes)
	if err != nil {
		return nil, err
	}
	if bytesRead != len(endBytes) || !bytes.Equal(endBytes, BinaryEndOfMetadata) {
		return nil, errors.New("the signature file appears to be corrupt")
	}
	pos += int64(bytesRead)

	s.ProgressReporter.ReportProgress("Reading signature", pos, inputLength)

	if hashAlgorithmStr != DefaultHashAlgorithm.Name() {
		return nil, fmt.Errorf("signature uses unsupported hash algorithm %s", hashAlgorithmStr)
	}
	hashAlgorithm := DefaultHashAlgorithm

	var rollingChecksum RollingChecksum
	switch rollingChecksumAlgorithmStr {
	case Adler32RollingChecksumName:
		rollingChecksum = NewAdler32RollingChecksum()
	case Adler32RollingChecksumV2Name:
		rollingChecksum = NewAdler32RollingChecksumV2()
	default:
		return nil, fmt.Errorf("signature uses unsupported rolling checksum algorithm %s", rollingChecksumAlgorithmStr)
	}

	expectedHashLength := hashAlgorithm.HashLength()
	remainingBytes := inputLength - pos
	signatureSize := 2 + 4 + expectedHashLength

	if remainingBytes%int64(signatureSize) != 0 {
		return nil, errors.New("the signature file appears to be corrupt; at least one chunk has data missing")
	}

	expectedNumberOfChunks := remainingBytes / int64(signatureSize)

	chunks := make([]*ChunkSignature, 0, expectedNumberOfChunks)

	chunkStart := int64(0)
	iter := NewReaderIteratorSize(input, signatureSize)
	for iter.Next() {
		block := iter.Current
		blockBytesRead := len(iter.Current)
		if blockBytesRead != signatureSize {
			return nil, fmt.Errorf("expecting to read %d bytes for ChunkSignature but only got %d", signatureSize, blockBytesRead)
		}
		pos += int64(blockBytesRead)

		length := uint16(block[0]) | uint16(block[1])<<8

		checksum := uint32(block[2]) | uint32(block[3])<<8 | uint32(block[4])<<16 | uint32(block[5])<<24

		chunks = append(chunks, &ChunkSignature{
			StartOffset:     chunkStart,
			Length:          length,
			RollingChecksum: checksum,
			Hash:            append([]byte(nil), block[6:]...), // copy the buffer as the next read around the loop is going to overwrite 'block'
		})

		chunkStart += int64(length)

		s.ProgressReporter.ReportProgress("Reading signature", pos, inputLength)
	}
	if err = iter.Err(); err != nil {
		return nil, err
	}

	return &Signature{
		HashAlgorithm:            hashAlgorithm,
		RollingChecksumAlgorithm: rollingChecksum,
		Chunks:                   chunks,
	}, nil
}
