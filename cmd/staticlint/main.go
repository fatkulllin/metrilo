package main

import (
	"strings"

	"github.com/gordonklaus/ineffassign/pkg/ineffassign"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/staticcheck"
)

func main() {

	var analyzers []*analysis.Analyzer

	for _, v := range staticcheck.Analyzers {
		if strings.HasPrefix(v.Analyzer.Name, "SA") {
			analyzers = append(analyzers, v.Analyzer)
		}
		if v.Analyzer.Name == "ST1000" {
			analyzers = append(analyzers, v.Analyzer)
		}
	}

	analyzers = append(analyzers, ineffassign.Analyzer)
	analyzers = append(analyzers, NoExitAnalyzer)
	analyzers = append(analyzers, printf.Analyzer, shadow.Analyzer, structtag.Analyzer)
	multichecker.Main(
		analyzers...,
	)

}
