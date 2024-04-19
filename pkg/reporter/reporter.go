package reporter

import "github.com/colussim/go-cloc/pkg/sorter"

type Reporter interface {
	GenerateReportByLanguage(summary *sorter.SortedSummary) error
	GenerateReportByFile(summary *sorter.SortedSummary) error
}
