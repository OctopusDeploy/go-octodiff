package octodiff

// TODO Obsolete: This is non standard implimentation of Adler32, Adler32RollingChecksumV2 should be used instead.

type Adler32RollingChecksum struct{}

const Adler32RollingChecksumName = "Adler32"

func NewAdler32RollingChecksum() *Adler32RollingChecksum {
	return &Adler32RollingChecksum{}
}

func (_ *Adler32RollingChecksum) Name() string {
	return Adler32RollingChecksumName
}

func (_ *Adler32RollingChecksum) Calculate(block []byte) uint32 {
	a := uint32(1)
	b := uint32(0)

	for _, z := range block {
		a = (uint32(z) + a) & 0xffff
		b = (b + a) & 0xffff
	}
	return (b << 16) | a
}

func (_ *Adler32RollingChecksum) Rotate(checksum uint32, remove byte, add byte, chunkSize int) uint32 {
	b := checksum >> 16 & 0xffff
	a := checksum & 0xffff

	a = (a - uint32(remove) + uint32(add)) & 0xffff
	b = (b - (uint32(chunkSize) * uint32(remove)) + a - 1) & 0xffff

	return (b << 16) | a
}

var _ RollingChecksum = (*Adler32RollingChecksum)(nil)
