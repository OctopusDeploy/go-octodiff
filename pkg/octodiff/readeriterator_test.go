package octodiff

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

type mockReader struct {
	Callbacks []func(p []byte) (n int, err error)
}

func newMockReader(Callbacks ...func(p []byte) (n int, err error)) *mockReader {
	return &mockReader{Callbacks: Callbacks}
}

func (m *mockReader) Read(p []byte) (n int, err error) {
	if len(m.Callbacks) == 0 {
		return 0, errors.New("test error: read past end of providers")
	}
	callback := m.Callbacks[0]
	m.Callbacks = m.Callbacks[1:]
	return callback(p)
}

func (m *mockReader) AllCallbacksConsumed() bool {
	return len(m.Callbacks) == 0
}

func returnData(data []byte) func(p []byte) (n int, err error) {
	return func(p []byte) (int, error) {
		return copy(p, data), nil
	}
}

func returnDataWithError(data []byte, err error) func(p []byte) (n int, err error) {
	return func(p []byte) (int, error) {
		return copy(p, data), err
	}
}

func returnError(err error) func(p []byte) (n int, err error) {
	return returnDataWithError(nil, err)
}

func returnDataWithEof(data []byte) func(p []byte) (n int, err error) {
	return returnDataWithError(data, io.EOF)
}

var returnEof = returnDataWithEof(nil)

func TestReaderIterator_OneShotExactBufferSize(t *testing.T) {
	reader := newMockReader(
		returnData([]byte("abcde")),
		returnEof)

	var received []byte
	iter := NewReaderIteratorSize(reader, 5)
	for iter.Next() {
		received = append(received, iter.Current...)
	}

	assert.Nil(t, iter.Err())
	assert.Equal(t, "abcde", string(received))

	assert.True(t, reader.AllCallbacksConsumed())
}

func TestReaderIterator_OneShotExactBufferSizeNBytes(t *testing.T) {
	reader := newMockReader(
		returnData([]byte("abcde"))) // Note there's no EOF here, the reader stops itself after reading N bytes

	var received []byte
	iter := NewReaderIteratorSizeNBytes(reader, 5, 5)
	for iter.Next() {
		received = append(received, iter.Current...)
	}

	assert.Nil(t, iter.Err())
	assert.Equal(t, "abcde", string(received))

	assert.True(t, reader.AllCallbacksConsumed())
}

func TestReaderIterator_OneShotExactBufferSizeNBytesTruncates(t *testing.T) {
	reader := newMockReader(
		returnData([]byte("abcde"))) // Note there's no EOF here, the reader stops itself after reading N bytes

	var received []byte
	iter := NewReaderIteratorSizeNBytes(reader, 5, 3)
	for iter.Next() {
		received = append(received, iter.Current...)
	}

	assert.Nil(t, iter.Err())
	assert.Equal(t, "abc", string(received))

	assert.True(t, reader.AllCallbacksConsumed())
}

func TestReaderIterator_OneShotExactBufferSizeInlineEof(t *testing.T) {
	reader := newMockReader(
		returnDataWithEof([]byte("abcde"))) // there's an EOF but it doesn't matter

	var received []byte
	iter := NewReaderIteratorSize(reader, 5)
	for iter.Next() {
		received = append(received, iter.Current...)
	}

	assert.Nil(t, iter.Err())
	assert.Equal(t, "abcde", string(received))

	assert.True(t, reader.AllCallbacksConsumed())
}

func TestReaderIterator_OneShotExactBufferSizeNBytesInlineEof(t *testing.T) {
	reader := newMockReader(
		returnDataWithEof([]byte("abcde")))

	var received []byte
	iter := NewReaderIteratorSizeNBytes(reader, 5, 5)
	for iter.Next() {
		received = append(received, iter.Current...)
	}

	assert.Nil(t, iter.Err())
	assert.Equal(t, "abcde", string(received))

	assert.True(t, reader.AllCallbacksConsumed())
}

func TestReaderIterator_OneShotExactBufferSizeNBytesInlineEofTruncates(t *testing.T) {
	reader := newMockReader(
		returnDataWithEof([]byte("abcde")))

	var received []byte
	iter := NewReaderIteratorSizeNBytes(reader, 5, 3)
	for iter.Next() {
		received = append(received, iter.Current...)
	}

	assert.Nil(t, iter.Err())
	assert.Equal(t, "abc", string(received))

	assert.True(t, reader.AllCallbacksConsumed())
}

func TestReaderIterator_MultiShotExactBufferSize(t *testing.T) {
	reader := newMockReader(
		returnData([]byte("abcde")),
		returnData([]byte("fghij")),
		returnData([]byte("klmno")),
		returnEof)

	var received []byte
	iter := NewReaderIteratorSize(reader, 5)
	for iter.Next() {
		received = append(received, iter.Current...)
	}

	assert.Nil(t, iter.Err())
	assert.Equal(t, "abcdefghijklmno", string(received))

	assert.True(t, reader.AllCallbacksConsumed())
}

func TestReaderIterator_MultiShotExactBufferSizeNBytes(t *testing.T) {
	reader := newMockReader(
		returnData([]byte("abcde")),
		returnData([]byte("fghij")),
		returnData([]byte("klmno"))) // EOF doesn't get read, the reader stops itself

	var received []byte
	iter := NewReaderIteratorSizeNBytes(reader, 5, 15)
	for iter.Next() {
		received = append(received, iter.Current...)
	}

	assert.Nil(t, iter.Err())
	assert.Equal(t, "abcdefghijklmno", string(received))

	assert.True(t, reader.AllCallbacksConsumed())
}

func TestReaderIterator_MultiShotExactBufferSizeNBytesTruncates(t *testing.T) {
	reader := newMockReader(
		returnData([]byte("abcde")),
		returnData([]byte("fghij")),
		returnData([]byte("klmno"))) // EOF doesn't get read, the reader stops itself

	var received []byte
	iter := NewReaderIteratorSizeNBytes(reader, 5, 12)
	for iter.Next() {
		received = append(received, iter.Current...)
	}

	assert.Nil(t, iter.Err())
	assert.Equal(t, "abcdefghijkl", string(received))

	assert.True(t, reader.AllCallbacksConsumed())
}

func TestReaderIterator_MultiShotExactBufferSizeInlineEof(t *testing.T) {
	reader := newMockReader(
		returnData([]byte("abcde")),
		returnData([]byte("fghij")),
		returnDataWithEof([]byte("klmno")))

	var received []byte
	iter := NewReaderIteratorSize(reader, 5)
	for iter.Next() {
		received = append(received, iter.Current...)
	}

	assert.Nil(t, iter.Err())
	assert.Equal(t, "abcdefghijklmno", string(received))

	assert.True(t, reader.AllCallbacksConsumed())
}

func TestReaderIterator_MultiShotExactBufferSizeNBytesInlineEof(t *testing.T) {
	reader := newMockReader(
		returnData([]byte("abcde")),
		returnData([]byte("fghij")),
		returnDataWithEof([]byte("klmno"))) // eof is there but doesn't matter

	var received []byte
	iter := NewReaderIteratorSizeNBytes(reader, 5, 15)
	for iter.Next() {
		received = append(received, iter.Current...)
	}

	assert.Nil(t, iter.Err())
	assert.Equal(t, "abcdefghijklmno", string(received))

	assert.True(t, reader.AllCallbacksConsumed())
}

func TestReaderIterator_OneShotLargerBufferSize(t *testing.T) {
	reader := newMockReader(
		returnData([]byte("abcde")),
		returnEof)

	var received []byte
	iter := NewReaderIteratorSize(reader, 500)
	for iter.Next() {
		received = append(received, iter.Current...)
	}

	assert.Nil(t, iter.Err())
	assert.Equal(t, "abcde", string(received))

	assert.True(t, reader.AllCallbacksConsumed())
}

func TestReaderIterator_OneShotLargerBufferSizeNBytes(t *testing.T) {
	reader := newMockReader(
		returnData([]byte("abcde"))) // we don't need the EOF because the reader stops itself early

	var received []byte
	iter := NewReaderIteratorSizeNBytes(reader, 500, 5)
	for iter.Next() {
		received = append(received, iter.Current...)
	}

	assert.Nil(t, iter.Err())
	assert.Equal(t, "abcde", string(received))

	assert.True(t, reader.AllCallbacksConsumed())
}

func TestReaderIterator_OneShotLargerBufferSizeInlineEof(t *testing.T) {
	reader := newMockReader(
		returnDataWithEof([]byte("abcde")))

	var received []byte
	iter := NewReaderIteratorSize(reader, 500)
	for iter.Next() {
		received = append(received, iter.Current...)
	}

	assert.Nil(t, iter.Err())
	assert.Equal(t, "abcde", string(received))

	assert.True(t, reader.AllCallbacksConsumed())
}

func TestReaderIterator_OneShotLargerBufferSizeNBytesInlineEof(t *testing.T) {
	reader := newMockReader(
		returnDataWithEof([]byte("abcde")))

	var received []byte
	iter := NewReaderIteratorSizeNBytes(reader, 500, 5)
	for iter.Next() {
		received = append(received, iter.Current...)
	}

	assert.Nil(t, iter.Err())
	assert.Equal(t, "abcde", string(received))

	assert.True(t, reader.AllCallbacksConsumed())
}

func TestReaderIterator_MultiShotLargerBufferSizeInlineEof(t *testing.T) {
	reader := newMockReader(
		returnData([]byte("abcde")),
		returnData([]byte("fghij")),
		returnDataWithEof([]byte("klmno")))

	var received []byte
	iter := NewReaderIteratorSize(reader, 500)
	for iter.Next() {
		received = append(received, iter.Current...)
	}

	assert.Nil(t, iter.Err())
	assert.Equal(t, "abcdefghijklmno", string(received))

	assert.True(t, reader.AllCallbacksConsumed())
}

func TestReaderIterator_MultiShotLargerBufferSizeNBytesInlineEof(t *testing.T) {
	reader := newMockReader(
		returnData([]byte("abcde")),
		returnData([]byte("fghij")),
		returnDataWithEof([]byte("klmno")))

	var received []byte
	iter := NewReaderIteratorSizeNBytes(reader, 500, 15)
	for iter.Next() {
		received = append(received, iter.Current...)
	}

	assert.Nil(t, iter.Err())
	assert.Equal(t, "abcdefghijklmno", string(received))

	assert.True(t, reader.AllCallbacksConsumed())
}

func TestReaderIterator_MultiShotLargerBufferSizeNBytesTruncatesInlineEof(t *testing.T) {
	reader := newMockReader(
		returnData([]byte("abcde")),
		returnData([]byte("fghij")),
		returnDataWithEof([]byte("klmno")))

	var received []byte
	iter := NewReaderIteratorSizeNBytes(reader, 500, 12)
	for iter.Next() {
		received = append(received, iter.Current...)
	}

	assert.Nil(t, iter.Err())
	assert.Equal(t, "abcdefghijkl", string(received))

	assert.True(t, reader.AllCallbacksConsumed())
}

func TestReaderIterator_DifferentBlockSizes(t *testing.T) {
	reader := newMockReader(
		returnData([]byte("abcde")),
		returnData([]byte("f")),
		returnData([]byte("ghijklmnop")),
		// Note deliberate zero byte block in the middle which io.Reader explicitly allows but discourages
		// https://github.com/golang/go/issues/27531
		returnData(make([]byte, 0)),
		returnData([]byte("qr")),
		returnEof)

	var received []byte
	iter := NewReaderIteratorSize(reader, 500)
	for iter.Next() {
		received = append(received, iter.Current...)
	}

	assert.Nil(t, iter.Err())
	assert.Equal(t, "abcdefghijklmnopqr", string(received))

	assert.True(t, reader.AllCallbacksConsumed())
}

func TestReaderIterator_DifferentBlockSizesNBytes(t *testing.T) {
	reader := newMockReader(
		returnData([]byte("abcde")),
		returnData([]byte("f")),
		returnData([]byte("ghijklmnop")),
		// Note deliberate zero byte block in the middle which io.Reader explicitly allows but discourages
		// https://github.com/golang/go/issues/27531
		returnData(make([]byte, 0)),
		returnData([]byte("qr"))) // doesn't read the EOF

	var received []byte
	iter := NewReaderIteratorSizeNBytes(reader, 500, 18)
	for iter.Next() {
		received = append(received, iter.Current...)
	}

	assert.Nil(t, iter.Err())
	assert.Equal(t, "abcdefghijklmnopqr", string(received))

	assert.True(t, reader.AllCallbacksConsumed())
}

func TestReaderIterator_DifferentBlockSizesNBytesTruncates(t *testing.T) {
	reader := newMockReader(
		returnData([]byte("abcde")),
		returnData([]byte("f")),
		returnData([]byte("ghijklmnop")),
		// Note deliberate zero byte block in the middle which io.Reader explicitly allows but discourages
		// https://github.com/golang/go/issues/27531
		returnData(make([]byte, 0)),
		returnData([]byte("qr"))) // doesn't read the EOF

	var received []byte
	iter := NewReaderIteratorSizeNBytes(reader, 500, 17)
	for iter.Next() {
		received = append(received, iter.Current...)
	}

	assert.Nil(t, iter.Err())
	assert.Equal(t, "abcdefghijklmnopq", string(received))

	assert.True(t, reader.AllCallbacksConsumed())
}

func TestReaderIterator_Failure(t *testing.T) {
	reader := newMockReader(
		returnError(errors.New("x")))

	var received []byte
	iter := NewReaderIteratorSize(reader, 500)
	for iter.Next() {
		received = append(received, iter.Current...)
	}

	assert.EqualError(t, iter.Err(), "x")
	assert.Equal(t, []byte(nil), received)

	assert.True(t, reader.AllCallbacksConsumed())
}

func TestReaderIterator_FailureAfterReading(t *testing.T) {
	reader := newMockReader(
		returnData([]byte("abcde")),
		returnData([]byte("f")),
		returnError(errors.New("x")))

	var received []byte
	iter := NewReaderIteratorSize(reader, 500)
	for iter.Next() {
		received = append(received, iter.Current...)
	}

	assert.EqualError(t, iter.Err(), "x")
	assert.Equal(t, []byte("abcdef"), received)

	assert.True(t, reader.AllCallbacksConsumed())
}

func TestReaderIterator_FailureAfterReadingNBytes(t *testing.T) {
	reader := newMockReader(
		returnData([]byte("abcde")),
		returnData([]byte("f")),
		returnError(errors.New("x")))

	var received []byte
	iter := NewReaderIteratorSizeNBytes(reader, 500, 6)
	for iter.Next() {
		received = append(received, iter.Current...)
	}

	// we stop iterating before we reach the error so we never see it
	assert.Nil(t, iter.Err())
	assert.Equal(t, []byte("abcdef"), received)

	assert.False(t, reader.AllCallbacksConsumed())
}

func TestReaderIterator_NextAfterFailure(t *testing.T) {
	reader := newMockReader(
		returnData([]byte("abcde")),
		returnData([]byte("f")),
		returnError(errors.New("x")))

	var received []byte
	iter := NewReaderIteratorSize(reader, 500)
	for iter.Next() {
		received = append(received, iter.Current...)
	}

	assert.EqualError(t, iter.Err(), "x")
	assert.Equal(t, []byte("abcdef"), received)

	assert.False(t, iter.Next())

	// doesn't impact the outcome
	assert.EqualError(t, iter.Err(), "x")
	assert.Equal(t, []byte("abcdef"), received)

	assert.True(t, reader.AllCallbacksConsumed())
}

func TestReaderIterator_FailureWhileAlsoProvidingData(t *testing.T) {
	// quoting https://pkg.go.dev/io#Reader
	// > Callers should always process the n > 0 bytes returned before considering the error err.
	// > Doing so correctly handles I/O errors that happen after reading some bytes and also both of the allowed EOF behaviors.
	reader := newMockReader(
		returnData([]byte("abcde")),
		returnData([]byte("f")),
		returnDataWithError([]byte("ghi"), errors.New("x")))

	var received []byte
	iter := NewReaderIteratorSize(reader, 500)
	for iter.Next() {
		received = append(received, iter.Current...)
	}

	assert.EqualError(t, iter.Err(), "x")
	assert.Equal(t, []byte("abcdefghi"), received)

	assert.True(t, reader.AllCallbacksConsumed())
}

func TestReaderIterator_FailureWhileAlsoProvidingDataNBytes(t *testing.T) {
	// quoting https://pkg.go.dev/io#Reader
	// > Callers should always process the n > 0 bytes returned before considering the error err.
	// > Doing so correctly handles I/O errors that happen after reading some bytes and also both of the allowed EOF behaviors.
	reader := newMockReader(
		returnData([]byte("abcde")),
		returnData([]byte("f")),
		returnDataWithError([]byte("ghi"), errors.New("x")))

	var received []byte
	iter := NewReaderIteratorSizeNBytes(reader, 500, 8)
	for iter.Next() {
		received = append(received, iter.Current...)
	}

	assert.EqualError(t, iter.Err(), "x")
	assert.Equal(t, []byte("abcdefgh"), received)

	assert.True(t, reader.AllCallbacksConsumed())
}
