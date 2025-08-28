package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var NoExitAnalyzer = &analysis.Analyzer{
	Name: "noexit",
	Doc:  "reports direct calls to os.Exit in main.main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	skip := false

	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if x, ok := sel.X.(*ast.Ident); ok &&
					x.Name == "testing" && sel.Sel.Name == "MainStart" {
					skip = true
					return false
				}
			}
			return true
		})

		if skip {
			return nil, nil
		}
	}

	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok &&
					ident.Name == "os" && sel.Sel.Name == "Exit" {
					pass.Reportf(call.Pos(), "os.Exit should not be called directly in main")
				}
			}
			return true
		})
	}

	return nil, nil
}
