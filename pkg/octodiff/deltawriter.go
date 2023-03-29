package octodiff

import "io"

type DeltaWriter interface {
	WriteMetadata(hashAlgorithm HashAlgorithm, expectedNewFileHash []byte) error
	WriteCopyCommand(offset int64, length int64) error
	WriteDataCommand(source io.ReadSeeker, offset int64, length int64) error

	// A DeltaWriter may "hold" the last CopyCommand, to allow merging sequential copy commands
	// Because of this, we need to tell the writer when it's done to flush any unwritten CopyCommand
	Flush() error
}
