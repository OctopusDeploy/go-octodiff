package octodiff

type Signature struct {
	HashAlgorithm            HashAlgorithm
	RollingChecksumAlgorithm RollingChecksum
	Chunks                   []*ChunkSignature
}

type ChunkSignature struct {
	// StartOffset is not written to disk, it's just calculated in-memory
	StartOffset int64
	// These fields are included in the binary file
	Length          uint16
	Hash            []byte
	RollingChecksum uint32
}
