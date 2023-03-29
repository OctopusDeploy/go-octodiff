package octodiff

import (
	"encoding/binary"
	"errors"
	"io"
)

const (
	SignatureMinimumChunkSize = 128
	SignatureDefaultChunkSize = 2048
	SignatureMaximumChunkSize = 31 * 1024
)

type SignatureBuilder struct {
	ChunkSize                int
	HashAlgorithm            HashAlgorithm    // must be non-null
	RollingChecksumAlgorithm RollingChecksum  // must be non-null
	ProgressReporter         ProgressReporter // must be non-null
}

func NewSignatureBuilder() *SignatureBuilder {
	return &SignatureBuilder{
		ChunkSize:                SignatureDefaultChunkSize,
		HashAlgorithm:            DefaultHashAlgorithm,
		RollingChecksumAlgorithm: DefaultChecksumAlgorithm,
		ProgressReporter:         NopProgressReporter(),
	}
}

func (s *SignatureBuilder) Build(input io.Reader, inputLength int64, output io.Writer) error {
	err := s.ensureValid()
	if err != nil {
		return err
	}
	err = s.writeMetadata(inputLength, output)
	if err != nil {
		return err
	}
	err = s.writeChunkSignatures(input, inputLength, output)
	if err != nil {
		return err
	}
	return nil
}

func (s *SignatureBuilder) ensureValid() error {
	if s.ChunkSize < SignatureMinimumChunkSize {
		return errors.New("SignatureBuilder ChunkSize is less than minimum allowed")
	}
	if s.ChunkSize > SignatureMaximumChunkSize {
		return errors.New("SignatureBuilder ChunkSize is greater than maximum allowed")
	}
	return nil
}

func (s *SignatureBuilder) writeMetadata(inputLength int64, output io.Writer) error {
	s.ProgressReporter.ReportProgress("Hashing file", 0, inputLength)

	_, err := output.Write(BinarySignatureHeader)
	if err != nil {
		return err
	}
	_, err = output.Write(BinaryVersion)
	if err != nil {
		return err
	}
	err = writeLengthPrefixedString(output, s.HashAlgorithm.Name())
	if err != nil {
		return err
	}
	err = writeLengthPrefixedString(output, s.RollingChecksumAlgorithm.Name())
	if err != nil {
		return err
	}
	_, err = output.Write(BinaryEndOfMetadata)
	if err != nil {
		return err
	}

	s.ProgressReporter.ReportProgress("Hashing file", inputLength, inputLength)
	return nil
}

func (s *SignatureBuilder) writeChunkSignatures(input io.Reader, inputLength int64, output io.Writer) error {
	checksumAlgorithm := s.RollingChecksumAlgorithm
	hashAlgorithm := s.HashAlgorithm

	s.ProgressReporter.ReportProgress("Building signatures", 0, inputLength)

	start := int64(0)
	iter := NewReaderIteratorSize(input, s.ChunkSize)
	for iter.Next() {
		err := writeChunk(output, iter.Current, hashAlgorithm.HashOverData(iter.Current), checksumAlgorithm.Calculate(iter.Current))
		if err != nil {
			return err
		}

		start += int64(len(iter.Current))
		s.ProgressReporter.ReportProgress("Building signatures", start, inputLength)
	}
	return iter.Err()
}

func writeChunk(output io.Writer, block []byte, hash []byte, rollingChecksum uint32) error {
	err := binary.Write(output, binary.LittleEndian, uint16(len(block)))
	if err != nil {
		return err
	}
	err = binary.Write(output, binary.LittleEndian, rollingChecksum)
	if err != nil {
		return err
	}
	_, err = output.Write(hash)
	return err
}
