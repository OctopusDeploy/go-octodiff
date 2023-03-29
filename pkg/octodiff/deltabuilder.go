package octodiff

import (
	"bytes"
	"io"
	"math"
	"sort"
)

type DeltaBuilder struct {
	ProgressReporter ProgressReporter
}

func NewDeltaBuilder() *DeltaBuilder {
	return &DeltaBuilder{
		ProgressReporter: NopProgressReporter(),
	}
}

// Build creates a new delta file, writing it out using `deltaWriter`
// confusing naming: "newFile" isn't a new file that we are creating, but rather an existing file which is
// "new" in that we haven't created a delta for it yet.
func (d *DeltaBuilder) Build(newFile io.ReadSeeker, newFileLength int64, signatureFile io.Reader, signatureFileLength int64, deltaWriter DeltaWriter) error {
	signatureReader := NewSignatureReader()
	signatureReader.ProgressReporter = d.ProgressReporter

	signature, err := signatureReader.ReadSignature(signatureFile, signatureFileLength)
	if err != nil {
		return err
	}

	chunks := signature.Chunks
	hash, err := signature.HashAlgorithm.HashOverReader(newFile)
	if err != nil {
		return err
	}
	_, err = newFile.Seek(0, io.SeekStart) // HashOverReader reads the entire newFile; we need to seek back to the start to process it
	if err != nil {
		return err
	}

	err = deltaWriter.WriteMetadata(signature.HashAlgorithm, hash)
	if err != nil {
		return err
	}

	sort.Slice(chunks, func(i, j int) bool {
		// aligns with C# ChunkSignatureChecksumComparer
		x, y := chunks[i], chunks[j]
		if x.RollingChecksum == y.RollingChecksum {
			return x.StartOffset < y.StartOffset
		}
		return x.RollingChecksum < y.RollingChecksum
	})

	chunkMap, minChunkSize, maxChunkSize := d.createChunkMap(chunks)

	lastMatchPosition := int64(0)
	buffer := make([]byte, defaultReadBufferSize)
	d.ProgressReporter.ReportProgress("Building delta", int64(0), newFileLength)

	startPosition := int64(0)

	for {
		bytesRead, fileReadErr := newFile.Read(buffer)
		if bytesRead > 0 { // we got some bytes, process them
			checksumAlgorithm := signature.RollingChecksumAlgorithm
			checksum := uint32(0)

			remainingPossibleChunkSize := maxChunkSize

			// slide a window over the buffer, looking for anything that matches our known list of chunks
			for i := 0; i < (bytesRead - minChunkSize + 1); i++ {
				readSoFar := startPosition + int64(i)

				remainingBytes := bytesRead - i
				if remainingBytes < maxChunkSize {
					remainingPossibleChunkSize = minChunkSize
				}

				if i == 0 || remainingBytes < maxChunkSize { // we are either at the start or end of buffer; calculate a full checksum
					checksum = checksumAlgorithm.Calculate(buffer[i : i+remainingPossibleChunkSize])
				} else { // we are stepping through the buffer, just rotate the existing checksum
					remove := buffer[i-1]
					add := buffer[i+remainingPossibleChunkSize-1]
					checksum = checksumAlgorithm.Rotate(checksum, remove, add, remainingPossibleChunkSize)
				}

				d.ProgressReporter.ReportProgress("Building delta", readSoFar, newFileLength)

				if readSoFar-(lastMatchPosition-int64(remainingPossibleChunkSize)) < int64(remainingPossibleChunkSize) {
					continue
				}

				startIndex, ok := chunkMap[checksum]
				if !ok {
					continue // we didn't match any known chunks. Skip, and the skipped data will be picked up later in a Data command based on lastMatchPosition
				}

				for j := startIndex; j < len(chunks) && chunks[j].RollingChecksum == checksum; j++ {
					chunk := chunks[j]

					sha := signature.HashAlgorithm.HashOverData(buffer[i : i+remainingPossibleChunkSize])

					if bytes.Equal(sha, chunk.Hash) {
						// we matched a chunk. Write any data in between it and the previous match as data, then write the 'copy' command for a chunk
						readSoFar = readSoFar + int64(remainingPossibleChunkSize)

						missing := readSoFar - lastMatchPosition
						if missing > int64(remainingPossibleChunkSize) {
							err = deltaWriter.WriteDataCommand(newFile, lastMatchPosition, missing-int64(remainingPossibleChunkSize))
							if err != nil {
								return err
							}
						}

						err = deltaWriter.WriteCopyCommand(chunk.StartOffset, int64(chunk.Length))
						if err != nil {
							return err
						}
						lastMatchPosition = readSoFar
						break
					}
				}
			}
		}
		if fileReadErr != nil {
			if fileReadErr == io.EOF { // all done
				break
			}
			return fileReadErr // something else went wrong after processing bytes, fail!
		}
		// If we didn't read a full buffer size, then assume we reached the end of newFile and exit the loop.
		// Note that Go's reader interface doesn't promise that it will always give you N bytes, even if N
		// bytes are available, in practice for file readers it does, and because we jump around seeking within
		// newFile there's not a great way of otherwise determining EOF without rewriting this whole thing
		if bytesRead < len(buffer) {
			break
		}

		// seek backwards by maxChunkSize+1, so we can read that stuff, and continue sliding the window over the file
		// note we mutate startPosition, so it is ready for the next time round the loop
		startPosition, err = newFile.Seek(-int64(maxChunkSize)+1, io.SeekCurrent)
		if err != nil {
			return err
		}
	}

	// we've reached the end of the file. Write any trailing data as a 'Data' command
	if newFileLength != lastMatchPosition {
		err = deltaWriter.WriteDataCommand(newFile, lastMatchPosition, newFileLength-lastMatchPosition)
		if err != nil {
			return err
		}
	}

	return deltaWriter.Flush()
}

// returns chunkMap, minChunkSize, maxChunkSize
func (d *DeltaBuilder) createChunkMap(chunks []*ChunkSignature) (map[uint32]int, int, int) {
	d.ProgressReporter.ReportProgress("Creating chunk map", 0, int64(len(chunks)))

	maxChunkSize := uint16(0)
	minChunkSize := uint16(math.MaxUint16)

	chunkMap := make(map[uint32]int)

	for chunkIdx, chunk := range chunks {
		if chunk.Length > maxChunkSize {
			maxChunkSize = chunk.Length
		}
		if chunk.Length < minChunkSize {
			minChunkSize = chunk.Length
		}

		if _, ok := chunkMap[chunk.RollingChecksum]; !ok {
			chunkMap[chunk.RollingChecksum] = chunkIdx
		}
		d.ProgressReporter.ReportProgress("Creating chunk map", int64(chunkIdx), int64(len(chunks)))
	}
	return chunkMap, int(minChunkSize), int(maxChunkSize)
}
