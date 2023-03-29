package delta

import (
	"bufio"
	"errors"
	"github.com/OctopusDeploy/go-octodiff/pkg/octodiff"
	"github.com/spf13/cobra"
	"io"
	"os"
)

type DeltaOptions struct {
	SignatureFile string
	NewFile       string
	DeltaFile     string
	Progress      bool
}

func NewCmdDelta() *cobra.Command {
	deltaOpts := &DeltaOptions{}
	cmd := &cobra.Command{
		Use:  "delta <signature-file> <new-file> [<delta-file>]",
		Long: "Given a signature file and a new file, creates a delta file",
		RunE: func(c *cobra.Command, args []string) error {
			// pick up positional arguments if not explicitly specified using --basis-file and --signature-file
			argOffset := 0
			if deltaOpts.SignatureFile == "" && len(args) > argOffset {
				deltaOpts.SignatureFile = args[argOffset]
				argOffset += 1
			}
			if deltaOpts.NewFile == "" && len(args) > argOffset {
				deltaOpts.NewFile = args[argOffset]
				argOffset += 1
			}
			if deltaOpts.DeltaFile == "" && len(args) > argOffset {
				deltaOpts.DeltaFile = args[argOffset]
			}

			return deltaRun(deltaOpts)
		},
	}

	flags := cmd.Flags()

	flags.StringVarP(&deltaOpts.SignatureFile, "signature-file", "", "", "The file containing the signature from the basis file.")
	flags.StringVarP(&deltaOpts.NewFile, "new-file", "", "", "The file to create the delta from.")
	flags.StringVarP(&deltaOpts.DeltaFile, "delta-file", "", "", "The file to write the delta to.")

	flags.BoolVarP(&deltaOpts.Progress, "progress", "", false, "Whether progress should be written to stdout")

	return cmd
}

func deltaRun(opts *DeltaOptions) error {
	signatureFilePath := opts.SignatureFile
	newFilePath := opts.NewFile
	deltaFilePath := opts.DeltaFile

	if signatureFilePath == "" {
		return errors.New("No signature file was specified")
	}
	if newFilePath == "" {
		return errors.New("No new file was specified")
	}

	signatureFile, err := os.Open(signatureFilePath)
	if errors.Is(err, os.ErrNotExist) {
		return errors.New("signature file does not exist or could not be opened")
	}
	if err != nil {
		return err
	}
	defer func() { _ = signatureFile.Close() }()
	signatureFileInfo, err := signatureFile.Stat()
	if err != nil {
		return err
	}

	newFile, err := os.Open(newFilePath)
	if errors.Is(err, os.ErrNotExist) {
		return errors.New("new file does not exist or could not be opened")
	}
	if err != nil {
		return err
	}
	defer func() { _ = newFile.Close() }()

	newFileInfo, err := newFile.Stat()
	if err != nil {
		return err
	}

	if deltaFilePath == "" {
		deltaFilePath = newFilePath + ".octodelta"
	} else {
		// mkdir_p on the signature file path directory? why?
	}

	deltaFile, err := os.Create(deltaFilePath)
	if err != nil {
		return err
	}
	defer func() { _ = deltaFile.Close() }()

	delta := octodiff.NewDeltaBuilder()
	if opts.Progress {
		delta.ProgressReporter = octodiff.NewStdoutProgressReporter()
	}

	// not using bufIo over newFile because we seek all over the place internally and bufio.Reader is not a ReadSeeker
	var signatureFileReader io.Reader = bufio.NewReaderSize(signatureFile, 4*1024*1024)
	var deltaFileWriter = bufio.NewWriter(deltaFile)
	err = delta.Build(newFile, newFileInfo.Size(), signatureFileReader, signatureFileInfo.Size(), octodiff.NewBinaryDeltaWriter(deltaFileWriter))
	if err != nil {
		return err
	}

	return deltaFileWriter.Flush()
}
