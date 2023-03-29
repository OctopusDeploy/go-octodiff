package octodiff

import (
	"encoding/binary"
	"io"
)

type BinaryDeltaWriter struct {
	Output             io.Writer
	bufferedCopyOffset int64
	bufferedCopyLength int64
}

var _ DeltaWriter = (*BinaryDeltaWriter)(nil)

func NewBinaryDeltaWriter(output io.Writer) *BinaryDeltaWriter {
	return &BinaryDeltaWriter{
		Output: output,
	}
}

func (w *BinaryDeltaWriter) WriteMetadata(hashAlgorithm HashAlgorithm, expectedNewFileHash []byte) error {
	_, err := w.Output.Write(BinaryDeltaHeader)
	if err != nil {
		return err
	}
	_, err = w.Output.Write(BinaryVersion)
	if err != nil {
		return err
	}
	err = writeLengthPrefixedString(w.Output, hashAlgorithm.Name())
	if err != nil {
		return err
	}
	err = binary.Write(w.Output, binary.LittleEndian, int32(len(expectedNewFileHash)))
	if err != nil {
		return err
	}
	_, err = w.Output.Write(expectedNewFileHash)
	if err != nil {
		return err
	}
	_, err = w.Output.Write(BinaryEndOfMetadata)
	return err
}

// WriteCopyCommand writes the "Copy Command" header to `output`
// followed by offset and length; There's no data
func (w *BinaryDeltaWriter) WriteCopyCommand(offset int64, length int64) error {
	if w.bufferedCopyLength == 0 { // just buffer it
		w.bufferedCopyOffset = offset
		w.bufferedCopyLength = length
	} else { // we have a buffered value, either merge or write the previous value and buffer this one
		if w.bufferedCopyOffset+w.bufferedCopyLength == offset { // merge
			w.bufferedCopyLength += length
		} else { // write previous and buffer this one
			err := writeCopyCommand(w.Output, w.bufferedCopyOffset, w.bufferedCopyLength)
			w.bufferedCopyOffset = offset
			w.bufferedCopyLength = length
			return err
		}
	}

	return nil
}

func writeCopyCommand(output io.Writer, offset, length int64) error {
	_, err := output.Write(BinaryCopyCommand)
	if err != nil {
		return err
	}
	err = binary.Write(output, binary.LittleEndian, offset)
	if err != nil {
		return err
	}
	return binary.Write(output, binary.LittleEndian, length)
}

func (w *BinaryDeltaWriter) Flush() error {
	if w.bufferedCopyLength != 0 {
		err := writeCopyCommand(w.Output, w.bufferedCopyOffset, w.bufferedCopyLength)
		w.bufferedCopyOffset = 0
		w.bufferedCopyLength = 0
		return err
	}

	return nil
}

// WriteDataCommand writes the "Data Command" header to `output`
// then proceeds to read `length` bytes from `source`, seeking to `offset` and write those to `output`
func (w *BinaryDeltaWriter) WriteDataCommand(source io.ReadSeeker, offset int64, length int64) (err error) {
	err = w.Flush()
	if err != nil {
		return
	}

	_, err = w.Output.Write(BinaryDataCommand)
	if err != nil {
		return
	}
	err = binary.Write(w.Output, binary.LittleEndian, length)
	if err != nil {
		return
	}

	var originalPosition int64
	originalPosition, err = source.Seek(0, io.SeekCurrent) // doing a no-op seek is how you find out the current position of a Go reader
	if err != nil {
		return
	}
	// we need to ensure we seek back to originalPosition before exiting the function.
	defer func() {
		_, seekBackErr := source.Seek(originalPosition, io.SeekStart)
		if seekBackErr != nil {
			err = seekBackErr // this causes writeDataCommand to return this error
		}
	}()

	_, err = source.Seek(offset, io.SeekStart)
	if err != nil {
		return
	}

	iter := NewReaderIteratorSizeNBytes(source, 1024*1024, length)
	for iter.Next() {
		_, err = w.Output.Write(iter.Current)
		if err != nil {
			return err
		}
	}
	return iter.Err()
}
