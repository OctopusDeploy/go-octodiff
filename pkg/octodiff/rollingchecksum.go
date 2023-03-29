package octodiff

type RollingChecksum interface {
	Name() string
	Calculate(block []byte) uint32
	Rotate(checksum uint32, remove byte, add byte, chunkSize int) uint32
}

var DefaultChecksumAlgorithm RollingChecksum = NewAdler32RollingChecksum()
