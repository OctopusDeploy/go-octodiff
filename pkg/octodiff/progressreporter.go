package octodiff

import "fmt"

type ProgressReporter interface {
	ReportProgress(operation string, currentPosition int64, total int64)
}

// ----------------------------------------------------------------------------

type nopProgressReporter struct {
}

func (n nopProgressReporter) ReportProgress(operation string, currentPosition int64, total int64) {
	// safely does nothing
}

var nopProgressReporterInstance = &nopProgressReporter{}

func NopProgressReporter() ProgressReporter {
	return nopProgressReporterInstance
}

// ----------------------------------------------------------------------------

type stdoutProgressReporter struct {
	CurrentOperation   string
	ProgressPercentage int
}

func (s *stdoutProgressReporter) ReportProgress(operation string, currentPosition int64, total int64) {
	percent := int(float64(currentPosition)/float64(total)*100.0 + 0.5)
	if s.CurrentOperation != operation {
		s.ProgressPercentage = -1
		s.CurrentOperation = operation
	}

	if s.ProgressPercentage != percent && percent%10 == 0 {
		s.ProgressPercentage = percent
		fmt.Printf("%v: %d%%\n", s.CurrentOperation, percent)
	}
}

func NewStdoutProgressReporter() ProgressReporter {
	return &stdoutProgressReporter{}
}
