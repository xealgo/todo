package scanner

import "context"

// ResultsGenerator interface used to define a contract for generating scan results.
type ResultsGenerator interface {
	Generate(context.Context, ScanReport) error
}
