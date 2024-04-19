package reporter

import "github.com/colussim/gcloc_m/pkg/sorter"

type Reporter interface {
	GenerateReportByLanguage(summary *sorter.SortedSummary) error
	GenerateReportByFile(summary *sorter.SortedSummary) error
}
