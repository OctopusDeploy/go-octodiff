package octodiff_test

import (
	"bytes"
	"encoding/hex"
	"github.com/OctopusDeploy/go-octodiff/pkg/octodiff"
	"github.com/stretchr/testify/assert"
	"testing"
)

func readSignature(input []byte) (*octodiff.Signature, error) {
	reader := octodiff.NewSignatureReader()
	return reader.ReadSignature(bytes.NewReader(input), int64(len(input)))
}

func assertChunk(t *testing.T, chunk *octodiff.ChunkSignature, startOffset int64, checksum uint32, length uint16, hashAsHexString string) {
	assert.Equal(t, startOffset, chunk.StartOffset)
	assert.Equal(t, checksum, chunk.RollingChecksum)
	assert.Equal(t, length, chunk.Length)
	assert.Equal(t, hashAsHexString, hex.EncodeToString(chunk.Hash))
}

func TestReadsStandardSignature(t *testing.T) {
	input, _ := hex.DecodeString("4f43544f5349470104534841310741646c657233323e3e3e0802f79fa2f0330bd06982d3b5dbda6c1a6ad16687a0cdb03c0d")

	s, err := readSignature(input)
	assert.Nil(t, err)
	assert.Equal(t, "SHA1", s.HashAlgorithm.Name())
	assert.Equal(t, "Adler32", s.RollingChecksumAlgorithm.Name())

	assert.Equal(t, 1, len(s.Chunks))
	assertChunk(t, s.Chunks[0], 0, 4037189623, 520, "330bd06982d3b5dbda6c1a6ad16687a0cdb03c0d")
}

func TestReadsStandardSignatureAdlerV2(t *testing.T) {
	input, _ := hex.DecodeString("4f43544f5349470104534841310941646c6572333256323e3e3e0802f79fe5f8330bd06982d3b5dbda6c1a6ad16687a0cdb03c0d")

	s, err := readSignature(input)
	assert.Nil(t, err)
	assert.Equal(t, "SHA1", s.HashAlgorithm.Name())
	assert.Equal(t, "Adler32V2", s.RollingChecksumAlgorithm.Name())

	assert.Equal(t, 1, len(s.Chunks))
	assertChunk(t, s.Chunks[0], 0, 4175798263, 520, "330bd06982d3b5dbda6c1a6ad16687a0cdb03c0d")
}

func TestReadsSmallChunkSizeSignature(t *testing.T) {
	input, _ := hex.DecodeString("4f43544f5349470104534841310741646c657233323e3e3e8000951f26e719f3978cb607e80a9aab3abbcac8bb1ecbcecf3e80001f18260f0f73196c2aa57877ee5e31291a59b5afca4493658000e035f42a42c4a73471dea3b9746e22dd93893fd8549f11bd8000dd2ff46b72e00e30ecae4c70ee07721d221a3b8a6d1847fa08008a02860c21d4023a8ba580ecdba742e7400aa40b6e449bb3")

	s, err := readSignature(input)
	assert.Nil(t, err)
	assert.Equal(t, "SHA1", s.HashAlgorithm.Name())
	assert.Equal(t, "Adler32", s.RollingChecksumAlgorithm.Name())

	assert.Equal(t, 5, len(s.Chunks))
	assertChunk(t, s.Chunks[0], 0, 3878035349, 128, "19f3978cb607e80a9aab3abbcac8bb1ecbcecf3e")
	assertChunk(t, s.Chunks[1], 128, 254154783, 128, "0f73196c2aa57877ee5e31291a59b5afca449365")
	assertChunk(t, s.Chunks[2], 256, 720647648, 128, "42c4a73471dea3b9746e22dd93893fd8549f11bd")
	assertChunk(t, s.Chunks[3], 384, 1811165149, 128, "72e00e30ecae4c70ee07721d221a3b8a6d1847fa")
	assertChunk(t, s.Chunks[4], 512, 210109066, 8, "21d4023a8ba580ecdba742e7400aa40b6e449bb3")
}

func TestReadsLargeChunkSizeSignatureOverLargeFile(t *testing.T) {
	input, _ := hex.DecodeString("4f43544f5349470104534841310741646c657233323e3e3e007cb823382f5470f51bab46eeb3913379e7b70a0d7329a9afce007cb5278ac69c31becd9bcd36f9afbd350ec15f4c437fd0cb67007c7a20e05ec605af9c2fd5a61b60f65600f5849f6ce1c53cf1001cac9ce9f194d25de18f219fa7832df14593cade50d8b0d2a2")

	s, err := readSignature(input)
	assert.Nil(t, err)
	assert.Equal(t, "SHA1", s.HashAlgorithm.Name())
	assert.Equal(t, "Adler32", s.RollingChecksumAlgorithm.Name())

	assert.Equal(t, 4, len(s.Chunks))
	assertChunk(t, s.Chunks[0], 0, 792208312, 31744, "5470f51bab46eeb3913379e7b70a0d7329a9afce")
	assertChunk(t, s.Chunks[1], 31744, 3330942901, 31744, "9c31becd9bcd36f9afbd350ec15f4c437fd0cb67")
	assertChunk(t, s.Chunks[2], 63488, 1591746682, 31744, "c605af9c2fd5a61b60f65600f5849f6ce1c53cf1")
	assertChunk(t, s.Chunks[3], 95232, 4058619052, 7168, "94d25de18f219fa7832df14593cade50d8b0d2a2")
}
