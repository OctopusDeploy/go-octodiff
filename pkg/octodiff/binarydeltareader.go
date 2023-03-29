package octodiff

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type DeltaReader interface {
	ExpectedHash() ([]byte, error)
	HashAlgorithm() (HashAlgorithm, error)

	// Apply reads the delta file.
	// This method will invoke the second func to copy the data from the original file multiple times if needed.
	// This method will also invoke first func to write data from the delta file to the destination file multiple times if needed.
	Apply(
		/*writeData*/ func([]byte) error,
		/*copyData*/ func(int64, int64) error,
	) error
}

type BinaryDeltaReader struct {
	input io.Reader

	expectedHash    []byte
	hashAlgorithm   HashAlgorithm
	hasReadMetadata bool

	ProgressReporter ProgressReporter
}

func NewBinaryDeltaReader(input io.Reader) *BinaryDeltaReader {
	return &BinaryDeltaReader{
		input:            input,
		ProgressReporter: NopProgressReporter(),
	}
}

func (b *BinaryDeltaReader) ExpectedHash() ([]byte, error) {
	err := b.ensureMetadata()
	if err != nil {
		return nil, err
	}
	return b.expectedHash, nil
}

func (b *BinaryDeltaReader) HashAlgorithm() (HashAlgorithm, error) {
	err := b.ensureMetadata()
	if err != nil {
		return nil, err
	}
	return b.hashAlgorithm, nil
}

func (b *BinaryDeltaReader) Apply(writeData func([]byte) error, copyData func(int64, int64) error) error {
	err := b.ensureMetadata()
	if err != nil {
		return err
	}

	buffer := make([]byte, defaultReadBufferSize)

	cmdTypeByte := make([]byte, 1)
	for {
		// we should not reach EOF when reading other expected bytes like EOF, but we
		// can rech it here once we've consumed all the commands in a file
		bytesRead, err := b.input.Read(cmdTypeByte)
		if err == io.EOF {
			return nil // all done, finished reading the file
		}
		if err != nil {
			return err
		}
		if bytesRead != len(cmdTypeByte) {
			return errors.New("could not read command type byte")
		}

		//b.ProgressReporter.ReportProgress("Applying delta", reader.BaseStream.Position, fileLength)

		if bytes.Equal(cmdTypeByte, BinaryCopyCommand) {
			var start, length int64
			err = binary.Read(b.input, binary.LittleEndian, &start)
			if err != nil {
				return err
			}
			err = binary.Read(b.input, binary.LittleEndian, &length)
			if err != nil {
				return err
			}
			err = copyData(start, length)
			if err != nil {
				return err
			}
			// loop round to read the next command
		} else if bytes.Equal(cmdTypeByte, BinaryDataCommand) {
			var length int64
			err = binary.Read(b.input, binary.LittleEndian, &length)
			if err != nil {
				return err
			}

			iter := NewReaderIteratorBufferNBytes(b.input, buffer, length)
			for iter.Next() {
				err = writeData(iter.Current)
				if err != nil {
					return err
				}
			}
			err = iter.Err()
			if err != nil {
				return err
			}
			// loop round to read the next command
		} else {
			return errors.New("unexpected cmd byte in delta file")
		}
	}
}

var _ DeltaReader = (*BinaryDeltaReader)(nil)

func (b *BinaryDeltaReader) ensureMetadata() error {
	if b.hasReadMetadata {
		return nil
	}

	headerBytes := make([]byte, len(BinaryDeltaHeader))
	bytesRead, err := b.input.Read(headerBytes)
	if err != nil {
		return err
	}
	if bytesRead != len(BinaryDeltaHeader) || !bytes.Equal(headerBytes, BinaryDeltaHeader) {
		return errors.New("the delta file appears to be corrupt")
	}

	var versionBytes = make([]byte, len(BinaryVersion))
	bytesRead, err = b.input.Read(versionBytes)
	if err != nil {
		return err
	}
	if bytesRead != len(BinaryVersion) || !bytes.Equal(versionBytes, BinaryVersion) {
		return errors.New("the delta file uses a newer file format than this program can handle")
	}

	hashAlgorithmName, _, err := readLengthPrefixedString(b.input)
	if err != nil {
		return err
	}
	if hashAlgorithmName != DefaultHashAlgorithm.Name() {
		return errors.New("the delta file uses an unsupported hashing algorithm")
	}
	hashAlgorithm := DefaultHashAlgorithm
	b.hashAlgorithm = hashAlgorithm

	var hashLength int32
	err = binary.Read(b.input, binary.LittleEndian, &hashLength)
	if err != nil {
		return err
	}
	if int(hashLength) != hashAlgorithm.HashLength() {
		return errors.New("the delta file contains an invalid hash length")
	}

	hashBytes := make([]byte, hashLength)
	bytesRead, err = b.input.Read(hashBytes)
	if err != nil {
		return err
	}
	if bytesRead != len(hashBytes) {
		return fmt.Errorf("the delta file appears to be corrupt; expecting hash length of %d but only read %d bytes", hashLength, len(hashBytes))
	}
	b.expectedHash = hashBytes

	endOfMetaBytes := make([]byte, len(BinaryEndOfMetadata))
	bytesRead, err = b.input.Read(endOfMetaBytes)
	if err != nil {
		return err
	}
	if bytesRead != len(BinaryEndOfMetadata) || !bytes.Equal(endOfMetaBytes, BinaryEndOfMetadata) {
		return errors.New("the delta file appears to be corrupt")
	}

	b.hasReadMetadata = true
	return nil
}
