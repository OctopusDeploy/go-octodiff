package octodiff

var BinarySignatureHeader = []byte("OCTOSIG")
var BinaryDeltaHeader = []byte("OCTODELTA")
var BinaryEndOfMetadata = []byte(">>>")

var BinaryCopyCommand = []byte{0x60}
var BinaryDataCommand = []byte{0x80}
var BinaryVersion = []byte{0x01}
