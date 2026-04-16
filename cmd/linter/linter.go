package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(Analyzer)
}

var Analyzer = &analysis.Analyzer{
	Name: "shortener_linter",
	Doc:  "Report the use of log.Fatal and/or os.Exit functions outside the main function of the main package and report the presence of the built-in panic function in the project code",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		var funcName string
		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {

			case *ast.FuncDecl:
				funcName = node.Name.Name
			case *ast.CallExpr:
				switch fun := node.Fun.(type) {
				case *ast.Ident:
					if fun.Name == "panic" {
						pass.Reportf(fun.Pos(), "panic usage detected")
					}
				case *ast.SelectorExpr:
					if xIdent, ok := fun.X.(*ast.Ident); ok {
						if pass.Pkg.Name() == "main" && funcName == "main" {
							return true
						}
						if xIdent.Name == "log" && fun.Sel.Name == "Fatal" {
							pass.Reportf(fun.Pos(), "log.Fatal usage detected")
						}
						if xIdent.Name == "os" && fun.Sel.Name == "Exit" {
							pass.Reportf(fun.Pos(), "os.Exit usage detected")
						}
					}
				}
			}

			return true
		})
	}
	return nil, nil
}
