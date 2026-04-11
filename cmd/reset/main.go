package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/tools/go/packages"
)

type Package struct {
	Name     string
	Imports  []string
	Path     string
	RStructs []ResetableStruct
}

type ResetableStruct struct {
	Name   string
	Fields []FieldInfo
}

type FieldInfo struct {
	ResetValue string
}

type Rfile struct {
	file *ast.File
	pkg  *packages.Package
	dir  string
}

var receiver = "rs"

const templateStr = `
package {{.Name}}

{{if .Imports}}
import (
{{range .Imports}}"{{.}}"
{{end}}
){{end}}

{{range .RStructs}}
func (rs *{{.Name}}) Reset() {
	if rs == nil {
        return
    }

	{{range .Fields}}
	{{.ResetValue}}{{end}} 
}
{{end}} 
`

var tmpl = template.Must(template.New("resetStruct").Parse(templateStr))

var structWithReset []string

var files []Rfile

func main() {
	var result []Package
	cfg := &packages.Config{
		Mode: packages.LoadSyntax,
		Dir:  ".",
	}

	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		panic(err)
	}

	seen := map[*ast.File]struct{}{}

	for _, pkg := range pkgs {
		for i, file := range pkg.Syntax {
			dir := filepath.Dir(pkg.CompiledGoFiles[i])
			ast.Inspect(file, func(n ast.Node) bool {
				gen, ok := n.(*ast.GenDecl)
				if !ok || gen.Tok != token.TYPE {
					return true
				}

				// проверяем комментарий НАД type block
				if gen.Doc == nil {
					return true
				}

				for _, c := range gen.Doc.List {
					if strings.Contains(strings.TrimSpace(c.Text), "generate:reset") {
						for _, spec := range gen.Specs {
							ts, ok := spec.(*ast.TypeSpec)
							if !ok {
								continue
							}
							if _, ok := seen[file]; !ok {
								seen[file] = struct{}{}
								rfile := Rfile{
									file: file,
									pkg:  pkg,
									dir:  dir,
								}
								files = append(files, rfile)
							}
							structWithReset = append(structWithReset, ts.Name.Name)
						}
						break
					}
				}
				return true
			})
		}
	}

	for _, file := range files {
		var imports []string
		var resultRs []ResetableStruct
		ast.Inspect(file.file, func(n ast.Node) bool {
			gen, ok := n.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				return true
			}

			// теперь разбираем структуры внутри блока
			for _, spec := range gen.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				st, ok := ts.Type.(*ast.StructType)
				if !ok {
					continue
				}

				var fields []FieldInfo

				for _, field := range st.Fields.List {
					fieldResetStr, imp := getResetExpr(file.pkg.TypesInfo, field.Type, receiver, field.Names[0].Name)
					if imp != "" {
						imports = append(imports, imp)
					}

					fields = append(fields, FieldInfo{
						ResetValue: fieldResetStr,
					})
				}

				if len(fields) > 0 {
					resultRs = append(resultRs, ResetableStruct{
						Name:   ts.Name.Name,
						Fields: fields,
					})
				}
			}

			return true
		})
		if len(resultRs) > 0 {
			result = append(result, Package{
				Name:     file.pkg.Name,
				Imports:  imports,
				Path:     file.dir,
				RStructs: resultRs,
			})
		}
	}

	buildResetFuncs(result)
}

func getResetExpr(info *types.Info, expr ast.Expr, receiver, fieldName string) (string, string) {
	switch t := expr.(type) {

	case *ast.Ident:
		switch t.Name {
		case "string":
			return fmt.Sprintf("%s.%s = \"\"", receiver, fieldName), ""
		case "bool":
			return fmt.Sprintf("%s.%s = false", receiver, fieldName), ""
		default:
			return fmt.Sprintf("%s.%s = 0", receiver, fieldName), ""
		}

	case *ast.ArrayType:
		// slice
		return fmt.Sprintf("%s.%s = %s.%s[:0]", receiver, fieldName, receiver, fieldName), ""

	case *ast.MapType:
		return fmt.Sprintf("clear(%s.%s)", receiver, fieldName), ""

	case *ast.StarExpr:
		// указатель
		code, imp := getResetExpr(info, t.X, receiver, fieldName)
		return code, imp

	case *ast.SelectorExpr:
		// например time.Time, context.Context и т.д.

		pkg, ok := t.X.(*ast.Ident)
		if !ok {
			return "// unknown selector", ""
		}

		fullType := pkg.Name + "." + t.Sel.Name
		obj := info.Uses[t.Sel]
		if obj == nil || obj.Pkg() == nil {
			return "", ""
		}

		fullImport := obj.Pkg().Path()

		switch fullImport {
		case "time":
			return fmt.Sprintf("%s.%s = %s{}", receiver, fieldName, fullType), fullImport

		default:
			if contains(structWithReset, t.Sel.Name) {
				return fmt.Sprintf("%s.%s.Reset()", receiver, fieldName), ""
			}

			switch obj.Type().Underlying().(type) {
			case *types.Interface:
				return fmt.Sprintf("%s.%s = nil", receiver, fieldName), ""
			case *types.Pointer:
				return fmt.Sprintf("%s.%s = nil", receiver, fieldName), ""
			case *types.Struct:
				return fmt.Sprintf("%s.%s = &%s{}", receiver, fieldName, fullType), fullImport
			}
		}

	default:
		return "// unsupported type", ""
	}
	return "", ""
}

func buildResetFuncs(result []Package) {
	for _, pkg := range result {
		var buf bytes.Buffer
		err := tmpl.Execute(&buf, pkg)
		if err != nil {
			panic(err)
		}

		// форматируем код
		bufFmt, err := format.Source(buf.Bytes())
		if err != nil {
			panic(err)
		}

		//fmt.Println(string(bufFmt))

		// записываем сгенерированный код в файл
		err = os.WriteFile(fmt.Sprintf("%s/reset.gen.go", pkg.Path), bufFmt, 0644)
		if err != nil {
			panic(err)
		}
	}
}

func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}
