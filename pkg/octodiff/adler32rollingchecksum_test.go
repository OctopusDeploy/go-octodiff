package octodiff_test

import (
	"github.com/OctopusDeploy/go-octodiff/pkg/octodiff"
	"github.com/OctopusDeploy/go-octodiff/pkg/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAdler32RollingChecksum_Name(t *testing.T) {
	c := &octodiff.Adler32RollingChecksum{}

	assert.Equal(t, "Adler32", c.Name())
}

func TestAdler32RollingChecksum_Calculate(t *testing.T) {
	c := &octodiff.Adler32RollingChecksum{}
	block := test.TestData()

	assert.Equal(t, uint32(2755533412), c.Calculate(block[:100]))
	assert.Equal(t, uint32(2888047271), c.Calculate(block[1:101]))
	assert.Equal(t, uint32(2476743237), c.Calculate(block[2:102]))
	assert.Equal(t, uint32(591925890), c.Calculate(block[93:193]))
	assert.Equal(t, uint32(4037189623), c.Calculate(block))

	largeBlock := test.GenerateTestData(100 * 1024)
	assert.Equal(t, uint32(2928347280), c.Calculate(largeBlock))
}

func TestAdler32RollingChecksum_Rotate(t *testing.T) {

	c := &octodiff.Adler32RollingChecksum{}
	assert.Equal(t, uint32(3209698067), c.Rotate(uint32(2755533412), 0, 0xAF, 8))
	assert.Equal(t, uint32(3209698067), c.Rotate(uint32(2755533412), 0, 0xAF, 16))
	assert.Equal(t, uint32(3209698067), c.Rotate(uint32(2755533412), 0, 0xAF, 24))
	assert.Equal(t, uint32(3209698067), c.Rotate(uint32(2755533412), 0, 0xAF, 32))

	assert.Equal(t, uint32(3577289570), c.Rotate(uint32(3209698067), 0xAF, 0xFE, 8))
	assert.Equal(t, uint32(3485539170), c.Rotate(uint32(3209698067), 0xAF, 0xFE, 16))
	assert.Equal(t, uint32(3393788770), c.Rotate(uint32(3209698067), 0xAF, 0xFE, 24))
	assert.Equal(t, uint32(3302038370), c.Rotate(uint32(3209698067), 0xAF, 0xFE, 32))
}
