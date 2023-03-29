package patch

import (
	"bufio"
	"errors"
	"github.com/OctopusDeploy/go-octodiff/pkg/octodiff"
	"github.com/spf13/cobra"
	"io"
	"os"
)

type PatchOptions struct {
	BasisFile        string
	DeltaFile        string
	NewFile          string
	Progress         bool
	SkipVerification bool
}

func NewCmdPatch() *cobra.Command {
	patchOpts := &PatchOptions{}
	cmd := &cobra.Command{
		Use:  "patch <basis-file> <delta-file> <new-file>",
		Long: "Given a basis file, and a delta, produces the new file.",
		RunE: func(c *cobra.Command, args []string) error {
			// pick up positional arguments if not explicitly specified using --basis-file and --signature-file
			argOffset := 0
			if patchOpts.BasisFile == "" && len(args) > argOffset {
				patchOpts.BasisFile = args[argOffset]
				argOffset += 1
			}
			if patchOpts.DeltaFile == "" && len(args) > argOffset {
				patchOpts.DeltaFile = args[argOffset]
				argOffset += 1
			}
			if patchOpts.NewFile == "" && len(args) > argOffset {
				patchOpts.NewFile = args[argOffset]
				argOffset += 1
			}
			return patchRun(patchOpts)
		},
	}

	flags := cmd.Flags()

	flags.StringVarP(&patchOpts.BasisFile, "basis-file", "", "", "The file that the delta was created for.")
	flags.StringVarP(&patchOpts.DeltaFile, "delta-file", "", "", "The delta to apply to the basis file.")
	flags.StringVarP(&patchOpts.NewFile, "new-file", "", "", "The file to write the result to.")
	flags.BoolVarP(&patchOpts.Progress, "progress", "", false, "Whether progress should be written to stdout.")
	flags.BoolVarP(&patchOpts.SkipVerification, "skip-verification", "", false, "Skip checking whether the basis file is the same as the file used to produce the signature that created the delta.")

	return cmd
}

func patchRun(opts *PatchOptions) error {
	// validate args
	basisFilePath := opts.BasisFile
	if basisFilePath == "" {
		return errors.New("no basis file was specified")
	}
	deltaFilePath := opts.DeltaFile
	if deltaFilePath == "" {
		return errors.New("no delta file was specified")
	}
	newFilePath := opts.NewFile
	if newFilePath == "" {
		return errors.New("no new file was specified")

	}
	// open files
	basisFile, err := os.Open(basisFilePath)
	if errors.Is(err, os.ErrNotExist) {
		return errors.New("basis file does not exist or could not be opened")
	}
	if err != nil {
		return err
	}
	defer func() { _ = basisFile.Close() }()

	deltaFile, err := os.Open(deltaFilePath)
	if errors.Is(err, os.ErrNotExist) {
		return errors.New("delta file does not exist or could not be opened")
	}
	if err != nil {
		return err
	}
	defer func() { _ = deltaFile.Close() }()

	var deltaFileStream io.Reader = bufio.NewReader(deltaFile)
	deltaReader := octodiff.NewBinaryDeltaReader(deltaFileStream)

	{ // nested lexical scope to contain the actual file writing, so we can ensure we flush/close properly
		newFile, err := os.Create(newFilePath)
		if err != nil {
			return err
		}
		// we can't buffer IO for basisFile because it seeks all over the place
		newFileOutputStream := bufio.NewWriter(newFile)

		err = octodiff.ApplyDelta(
			basisFile,
			deltaReader,
			newFileOutputStream)

		flushErr := newFileOutputStream.Flush()
		_ = newFile.Close()
		if flushErr != nil {
			return flushErr
		}
		if err != nil {
			return err
		}
	}

	if opts.SkipVerification {
		return nil
	}

	// re-open the file to verify the hash

	newFileRead, err := os.Open(newFilePath)
	if errors.Is(err, os.ErrNotExist) {
		return errors.New("new file does not exist or could not be opened for verification")
	}
	if err != nil {
		return err
	}
	defer func() { _ = newFileRead.Close() }()
	newFileReadStream := bufio.NewReaderSize(newFileRead, 4*1024*1024)
	return octodiff.VerifyNewFile(newFileReadStream, deltaReader)
}
