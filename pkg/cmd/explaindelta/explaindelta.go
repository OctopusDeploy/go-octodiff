package explaindelta

import (
	"bufio"
	"encoding/hex"
	"errors"
	"github.com/OctopusDeploy/go-octodiff/pkg/octodiff"
	"github.com/spf13/cobra"
	"io"
	"os"
)

type ExplainDeltaOptions struct {
	DeltaFile string
}

func NewCmdExplainDelta() *cobra.Command {
	deltaOpts := &ExplainDeltaOptions{}
	cmd := &cobra.Command{
		Use:  "explain-delta <delta-file>",
		Long: "Prints instructions from a delta file; useful when debugging.",
		RunE: func(c *cobra.Command, args []string) error {
			// pick up positional arguments if not explicitly specified using --basis-file and --signature-file
			argOffset := 0
			if deltaOpts.DeltaFile == "" && len(args) > argOffset {
				deltaOpts.DeltaFile = args[argOffset]
				argOffset += 1
			}
			return explainDeltaRun(c, deltaOpts)
		},
	}

	flags := cmd.Flags()

	flags.StringVarP(&deltaOpts.DeltaFile, "delta-file", "", "", "The file to explain.")

	return cmd
}

func explainDeltaRun(cmd *cobra.Command, opts *ExplainDeltaOptions) error {
	deltaFilePath := opts.DeltaFile

	if deltaFilePath == "" {
		return errors.New("no delta file was specified")
	}

	deltaFile, err := os.Open(deltaFilePath)
	if errors.Is(err, os.ErrNotExist) {
		return errors.New("delta file does not exist or could not be opened")
	}
	if err != nil {
		return err
	}
	defer func() { _ = deltaFile.Close() }()
	//deltaFileInfo, err := deltaFile.Stat()
	//if err != nil {
	//	return err
	//}

	var deltaFileReader io.Reader = bufio.NewReaderSize(deltaFile, 4*1024*1024)
	deltaReader := octodiff.NewBinaryDeltaReader(deltaFileReader)

	return deltaReader.Apply(func(bytes []byte) error {
		if len(bytes) > 20 {
			cmd.Printf("Data: (%v bytes): {%v}...\n", len(bytes), hex.EncodeToString(bytes[:20]))
		} else {
			cmd.Printf("Data: (%v bytes): {%v}\n", len(bytes), hex.EncodeToString(bytes))
		}
		return nil
	}, func(start int64, length int64) error {
		cmd.Printf("Copy: %d bytes from offset %X\n", length, start)
		return nil
	})
}
