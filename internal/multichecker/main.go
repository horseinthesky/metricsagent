// multichecker is a custom analysis tool.
//
// It consists of:
//   - default go analysis checkers: lostcancel, printf, structtag, unreachable
//   - all staticcheck.io SA(staticcheck) checkers
//   - ST1001 staticcheck.io stylecheck checker
//   - S1001 staticcheck.io simple checker
//   - nilerr checker from https://github.com/gostaticanalysis/nilerr
//   - unuseparam checker finds a unused parameter but its name is not _
//   - custom checker which checks if you are using os.Exit in main function
package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/unreachable"

	"github.com/gostaticanalysis/nilerr"
	"github.com/gostaticanalysis/unuseparam"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/stylecheck"
)

func main() {
	var analyzers []*analysis.Analyzer
	// for _, analyser := range staticcheck.Analyzers {
	// 	analyzers = append(analyzers, analyser)
	// }

	analyzers = append(
		analyzers,
		lostcancel.Analyzer,
		printf.Analyzer,
		structtag.Analyzer,
		unreachable.Analyzer,
		stylecheck.Analyzers["ST1001"],
		simple.Analyzers["S1001"],
		nilerr.Analyzer,
		unuseparam.Analyzer,
		NoOSExitAnalyzer,
	)

	multichecker.Main(analyzers...)
}
