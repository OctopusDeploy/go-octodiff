package octodiff

import (
	"bytes"
	"errors"
	"io"
)

// ApplyDelta builds thew new file.
// Verifying the hash of the written file is done seperately, to allow the caller to use
// a buffered output writer to improve performance.
func ApplyDelta(basisFile io.ReadSeeker, deltaReader DeltaReader, output io.Writer) error {
	buffer := make([]byte, defaultReadBufferSize)

	return deltaReader.Apply(
		func(bytes []byte) error {
			_, err := output.Write(bytes)
			return err
		},
		func(offset int64, length int64) error {
			_, err := basisFile.Seek(offset, io.SeekStart)
			if err != nil {
				return err
			}

			iter := NewReaderIteratorBufferNBytes(basisFile, buffer, length)
			for iter.Next() {
				_, err = output.Write(iter.Current)
				if err != nil {
					return err
				}
			}
			return iter.Err()
		})
}

func VerifyNewFile(newFile io.Reader, deltaReader DeltaReader) error {
	sourceFileHash, err := deltaReader.ExpectedHash()
	if err != nil {
		return err
	}
	algorithm, err := deltaReader.HashAlgorithm()
	if err != nil {
		return err
	}

	actualHash, err := algorithm.HashOverReader(newFile)
	if err != nil {
		return err
	}

	if !bytes.Equal(sourceFileHash, actualHash) {
		return errors.New("verification of the patched file failed. The SHA1 hash of the patch result file, and the file that was used as input for the delta, do not match. This can happen if the basis file changed since the signatures were calculated")
	}
	return nil
}
