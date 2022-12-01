package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var NoOSExitFilteredAnalyzer = &analysis.Analyzer{
	Name:     "noosexit",
	Doc:      "check for direct os.Exit usage",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      runFiltered,
}

func runFiltered(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Here we filter only function deslaration nodes
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	inspect.Preorder(nodeFilter, func(node ast.Node) {
		fdecl, ok := node.(*ast.FuncDecl)
		if !ok {
			return
		}

		if fdecl.Name.Name != "main" {
			return
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
	})

	return nil, nil
}
