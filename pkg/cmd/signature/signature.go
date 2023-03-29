package signature

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/OctopusDeploy/go-octodiff/pkg/octodiff"
	"github.com/spf13/cobra"
	"io"
	"os"
)

type SignatureOptions struct {
	BasisFile     string
	SignatureFile string
	ChunkSize     int
	Progress      bool
}

func NewCmdSignature() *cobra.Command {
	signatureOpts := &SignatureOptions{}
	cmd := &cobra.Command{
		Use:     "signature <basis-file> [<signature-file>]",
		Long:    "Given a basis file, creates a signature file",
		Aliases: []string{"sig"},
		RunE: func(c *cobra.Command, args []string) error {
			// pick up positional arguments if not explicitly specified using --basis-file and --signature-file
			argOffset := 0
			if signatureOpts.BasisFile == "" && len(args) > argOffset {
				signatureOpts.BasisFile = args[argOffset]
				argOffset += 1
			}
			if signatureOpts.SignatureFile == "" && len(args) > argOffset {
				signatureOpts.SignatureFile = args[argOffset]
			}

			return signatureRun(signatureOpts)
		},
	}

	flags := cmd.Flags()

	flags.StringVarP(&signatureOpts.BasisFile, "basis-file", "f", "", "The file to read and create a signature from.")
	flags.StringVarP(&signatureOpts.SignatureFile, "signature-file", "o", "", "The file to write the signature to.")

	flags.IntVarP(&signatureOpts.ChunkSize, "chunk-size", "", octodiff.SignatureDefaultChunkSize,
		fmt.Sprintf("Maximum bytes per chunk. Defaults to %d. Min of %d, max of %d.",
			octodiff.SignatureDefaultChunkSize, octodiff.SignatureMinimumChunkSize, octodiff.SignatureMaximumChunkSize))

	flags.BoolVarP(&signatureOpts.Progress, "progress", "", false, "Whether progress should be written to stdout")

	return cmd
}

func signatureRun(opts *SignatureOptions) error {
	basisFilePath := opts.BasisFile
	signatureFilePath := opts.SignatureFile

	if basisFilePath == "" {
		return errors.New("No basis file was specified")
	}

	basisFile, err := os.Open(basisFilePath)
	if errors.Is(err, os.ErrNotExist) {
		return errors.New("basis file does not exist or could not be opened")
	}
	if err != nil {
		return err
	}
	defer func() { _ = basisFile.Close() }()

	basisFileInfo, err := basisFile.Stat()
	if err != nil {
		return err
	}

	if signatureFilePath == "" {
		signatureFilePath = basisFilePath + ".octosig"
	} else {
		// mkdir_p on the signature file path directory? why?
	}

	signatureFile, err := os.Create(signatureFilePath)
	if err != nil {
		return err
	}
	defer func() { _ = signatureFile.Close() }()

	signatureBuilder := octodiff.NewSignatureBuilder()
	if opts.Progress {
		signatureBuilder.ProgressReporter = octodiff.NewStdoutProgressReporter()
	}

	// For a 4.5 gb ISO file on my dev laptop (March 2023) C# octodiff takes 16 seconds to generate a signature.
	//
	// With a 4MB read buffer we take 8 seconds; with a default 4k buffer we take 9.5 seconds; without bufio we take 12 seconds
	// bufio on the writer is even more important. The above 8-second signature generation takes 40 seconds without it, but unlike the reader, write buffer size doesn't affect things noticeably
	var basisFileReader io.Reader = bufio.NewReaderSize(basisFile, 4*1024*1024)
	var signatureFileWriter = bufio.NewWriter(signatureFile)
	err = signatureBuilder.Build(basisFileReader, basisFileInfo.Size(), signatureFileWriter)
	if err != nil {
		return err
	}
	return signatureFileWriter.Flush()
}
