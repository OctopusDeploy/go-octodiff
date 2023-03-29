package octodiff

import "io"

// ReaderIterator lets you treat reading bytes from an io.Reader as a for-loop.
// The logic of "read all the byes from an io.Reader" is surprisingly complex in Go.
// this struct wraps it up into a "Next/Current" iterator style object as used by bufio.Scanner, or IEnumerator in C#
// Important: Like bufio.Scanner, if Next() returns false you MUST check Err() to see if it failed
// Important: Calling Next() mutates the io.Reader so you can't create more than one ReaderIterator per reader
// Note: This is designed to be stack-allocated by the caller, so the New functions don't return pointers
type ReaderIterator struct {
	// critical fields, must allocate with NewReaderIterator
	reader       io.Reader
	buffer       []byte
	nBytesToRead int64

	// progress fields, zero-init is good
	nBytesReadSoFar int64
	isCompleted     bool
	err             error

	// Output; zero-init is good
	Current []byte
}

func (b *ReaderIterator) Err() error {
	return b.err
}

// Next calls `Read` on the underlying reader, returning true
func (b *ReaderIterator) Next() bool {
	if b.isCompleted {
		return false // already completed
	}

	localReadBuffer := b.buffer

	if b.nBytesToRead > 0 { // if we've been told to stop after a certain number of bytes, control this by varying the buffer passed to Read
		bytesRemaining := b.nBytesToRead - b.nBytesReadSoFar
		if bytesRemaining < int64(len(b.buffer)) {
			localReadBuffer = b.buffer[:bytesRemaining]
		}
	}

	bytesRead, err := b.reader.Read(localReadBuffer)
	b.nBytesReadSoFar += int64(bytesRead)

	if b.nBytesToRead > 0 { // if we've been told to stop after a certain number of bytes, control this by simulating an EOF
		if b.nBytesReadSoFar >= b.nBytesToRead && err == nil { // don't squash an underlying error though
			err = io.EOF
		}
	}

	if err != nil {
		// last block. May or may not have data depending on underlying reader
		b.isCompleted = true
		if err != io.EOF {
			b.err = err
		}
	}
	// even if an error was returned (whether EOF or not), the reader can still provide data
	if bytesRead == len(localReadBuffer) { // don't slice the buffer if we read the whole thing
		b.Current = localReadBuffer
	} else {
		b.Current = localReadBuffer[:bytesRead]
	}
	// if we hit the last block AND there's no data to return, tell the caller we're done
	return bytesRead > 0 || !b.isCompleted
}

// NewReaderIteratorSize creates an iterator, allocating a buffer of `bufferSize`
func NewReaderIteratorSize(reader io.Reader, bufferSize int) ReaderIterator {
	return NewReaderIteratorBuffer(reader, make([]byte, bufferSize))
}

// NewReaderIteratorSize creates an iterator, allocating a buffer of `bufferSize`
func NewReaderIteratorSizeNBytes(reader io.Reader, bufferSize int, nBytesToRead int64) ReaderIterator {
	return NewReaderIteratorBufferNBytes(reader, make([]byte, bufferSize), nBytesToRead)
}

// NewReaderIteratorBuffer creates an iterator, referencing an already-allocated buffer that will read until EOF
func NewReaderIteratorBuffer(reader io.Reader, buffer []byte) ReaderIterator {
	return NewReaderIteratorBufferNBytes(reader, buffer, -1)
}

// NewReaderIteratorBufferNBytes creates an iterator that will stop after reading `nBytesToRead` bytes, referencing an already-allocated buffer
func NewReaderIteratorBufferNBytes(reader io.Reader, buffer []byte, nBytesToRead int64) ReaderIterator {
	return ReaderIterator{
		reader:       reader,
		buffer:       buffer,
		nBytesToRead: nBytesToRead,
		isCompleted:  false,
		err:          nil,
		Current:      nil,
	}
}
