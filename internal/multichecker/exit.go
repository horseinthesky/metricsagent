package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var NoOSExitAnalyzer = &analysis.Analyzer{
	Name: "noosexit",
	Doc:  "check for direct os.Exit usage",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		// Ignore all files not in main package
		if file.Name.Name != "main" {
			continue
		}

		ast.Inspect(file, func(node ast.Node) bool {
			// Ignore all nodes that are not fucntion desclarations
			fdecl, ok := node.(*ast.FuncDecl)
			if !ok {
				return true
			}

			// Ignore all function declarations that are not "main"
			if fdecl.Name.Name != "main" {
				return true
			}

			for _, stmt := range fdecl.Body.List {
				exprStmt, ok := stmt.(*ast.ExprStmt)
				if !ok {
					continue
				}

				callExpr, ok := exprStmt.X.(*ast.CallExpr)
				if !ok {
					continue
				}

				funcExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
				if !ok {
					continue
				}

				identifier, ok := funcExpr.X.(*ast.Ident)
				if !ok {
					continue
				}

				if identifier.Name == "os" && funcExpr.Sel.Name == "Exit" {
					pass.Reportf(identifier.Pos(), "usage of os.Exit in main func of main package is forbiden")
				}
			}

			return true
		})
	}

	return nil, nil
}
