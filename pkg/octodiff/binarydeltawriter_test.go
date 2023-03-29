package octodiff_test

import (
	"bytes"
	"encoding/hex"
	"github.com/OctopusDeploy/go-octodiff/pkg/octodiff"
	"github.com/OctopusDeploy/go-octodiff/pkg/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWritesHeader(t *testing.T) {
	b := bytes.NewBuffer(nil)
	w := octodiff.NewBinaryDeltaWriter(b)

	err := w.WriteMetadata(&octodiff.Sha1HashAlgorithm{}, test.GenerateTestData(20))
	assert.Nil(t, err)
	assert.Equal(t, "4f43544f44454c54410104534841311400000030820204308201aba003020102021418d83f07713e3e3e", hex.EncodeToString(b.Bytes()))
}

func TestWritesCopyCommand(t *testing.T) {
	b := bytes.NewBuffer(nil)
	w := octodiff.NewBinaryDeltaWriter(b)

	err := w.WriteCopyCommand(315412, 9874563)
	assert.Nil(t, err)

	err = w.Flush()
	assert.Nil(t, err)

	assert.Equal(t, "6014d004000000000083ac960000000000", hex.EncodeToString(b.Bytes()))
}

func TestWritesDataCommand(t *testing.T) {
	b := bytes.NewBuffer(nil)
	w := octodiff.NewBinaryDeltaWriter(b)

	source := bytes.NewReader(test.GenerateTestData(1024))

	err := w.WriteDataCommand(source, 337, 515)
	assert.Nil(t, err)
	assert.Equal(t, "8003020000000000000931ecf7f3bd4bce212cfd2cbaa3533051301d0603551d0e04160414badd278a31e012776afbfda4ead8fdce904f0efc301f0603551d23041830168014badd278a31e012776afbfda4ead8fdce904f0efc300f0603551d130101ff040530030101ff300a06082a8648ce3d04030203470030440220599cef920115b64a7d0bc7de55a84bba7f05ee78b9e903af7cb52b4a5dcc8ea2022006575445dab9c21325a48de3bd7ce51a34612015a74648787c7a7e032645377030820204308201aba003020102021418d83f07718be4121df0a18d7610faf8d7a3bec4300a06082a8648ce3d0403023058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c0454455354301e170d3233303332303039343834325a170d3234303331393039343834325a3058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f79310c300a060355040b0c03522644310d300b06035504030c04544553543059301306072a8648ce3d020106082a8648ce3d03010703420004504b77248d83e2e3e209bbb2297a0e4d24ff45e79eff88dd165e6419ae98512dabd2219da46e93d7ff98d5a1cb80", hex.EncodeToString(b.Bytes()))
}

func TestMergesSequentialCopyCommands(t *testing.T) {
	b := bytes.NewBuffer(nil)
	w := octodiff.NewBinaryDeltaWriter(b)

	var err error
	// these 3 get merged
	err = w.WriteCopyCommand(0, 128)
	assert.Nil(t, err)

	err = w.WriteCopyCommand(128, 128)
	assert.Nil(t, err)

	err = w.WriteCopyCommand(256, 128)
	assert.Nil(t, err)

	// this one doesn't because there's a 1-byte gap
	err = w.WriteCopyCommand(385, 128)
	assert.Nil(t, err)

	err = w.Flush()
	assert.Nil(t, err)

	// 0x60 signifies a copy command, we can see there's only two here
	assert.Equal(t, "60000000000000000080010000000000006081010000000000008000000000000000", hex.EncodeToString(b.Bytes()))
}

func TestDataCommandFlushesCopyCommand(t *testing.T) {
	b := bytes.NewBuffer(nil)
	w := octodiff.NewBinaryDeltaWriter(b)

	source := bytes.NewReader(test.GenerateTestData(1024))
	var err error
	// these would get merged but they won't because there's a Data command in the middle
	err = w.WriteCopyCommand(0, 128)
	assert.Nil(t, err)

	err = w.WriteDataCommand(source, 500, 128)
	assert.Nil(t, err)

	err = w.WriteCopyCommand(128, 128)
	assert.Nil(t, err)

	err = w.Flush()
	assert.Nil(t, err)

	assert.Equal(t, "6000000000000000008000000000000000808000000000000000bd7ce51a34612015a74648787c7a7e032645377030820204308201aba003020102021418d83f07718be4121df0a18d7610faf8d7a3bec4300a06082a8648ce3d0403023058310b30090603550406130241553113301106035504080c0a536f6d652d537461746531173015060355040a0c0e4f63746f707573204465706c6f796080000000000000008000000000000000", hex.EncodeToString(b.Bytes()))
}
