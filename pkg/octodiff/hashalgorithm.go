package octodiff

import (
	"crypto/sha1"
	"io"
)

type HashAlgorithm interface {
	Name() string
	HashLength() int
	HashOverData(data []byte) []byte
	HashOverReader(reader io.Reader) ([]byte, error)
}

// the only hash algorithm octodiff seems to use is sha1

type Sha1HashAlgorithm struct {
}

func (s *Sha1HashAlgorithm) Name() string {
	return "SHA1"
}

// returns the length in bytes of a Hash computed with this aglorithm
func (s *Sha1HashAlgorithm) HashLength() int {
	return sha1.Size
}

func (s *Sha1HashAlgorithm) HashOverData(data []byte) []byte {
	h := sha1.Sum(data)
	return h[:] // convert from fixed-length array to slice
}

// This will issue lots of 1k reads into the reader.
// It's up to the caller to pass us a bufio if performance is of concern
func (s *Sha1HashAlgorithm) HashOverReader(reader io.Reader) ([]byte, error) {
	sha := sha1.New()

	iter := NewReaderIteratorSize(reader, 1024)
	for iter.Next() {
		_, err := sha.Write(iter.Current)
		if err != nil {
			return nil, err
		}
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}
	return sha.Sum(nil), nil
}

var DefaultHashAlgorithm HashAlgorithm = &Sha1HashAlgorithm{}
