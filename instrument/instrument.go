package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
)

func instrumentPackage(filePath string) error {
	// 解析包源代码
	fset := token.NewFileSet()
	//pkgs, err := parser.ParseDir(fset, filePath, nil, parser.ParseComments)
	f, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	// 遍历每个函数定义,并插入instrumentation代码
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			instrumentFunction(x)
		}
		return true
	})

	// 格式化并写回文件
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, f); err != nil {
		return err
	}
	if err := os.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
		return err
	}

	return nil
}

func instrumentFunction(f *ast.FuncDecl) {
	// 在函数开头插入打印参数的代码
	printArgsStmt := &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "fmt"},
				Sel: &ast.Ident{Name: "Printf"},
			},
			Args: []ast.Expr{
				&ast.BasicLit{Kind: token.STRING, Value: `"Calling %s with arguments: %v\n"`},
				&ast.Ident{Name: "\"" + f.Name.Name + "\""},
				&ast.Ident{Name: f.Type.Params.List[0].Names[0].Name},
				&ast.Ident{Name: f.Type.Params.List[0].Names[1].Name},
			},
		},
	}
	f.Body.List = append([]ast.Stmt{printArgsStmt}, f.Body.List...)

	// 在函数返回前插入打印返回值的代码
	printReturnsStmt := &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{Name: "fmt"},
				Sel: &ast.Ident{Name: "Printf"},
			},
			Args: []ast.Expr{
				&ast.BasicLit{Kind: token.STRING, Value: `"Function %s returned: %v\n"`},
				&ast.Ident{Name: f.Name.Name},
				&ast.Ident{Name: "f.Results"},
			},
		},
	}
	if f.Type.Results != nil && len(f.Type.Results.List) > 0 {
		lastStmt := f.Body.List[len(f.Body.List)-1]
		if returnStmt, ok := lastStmt.(*ast.ReturnStmt); ok {
			f.Body.List = append(f.Body.List[:len(f.Body.List)-1], &ast.BlockStmt{
				List: append([]ast.Stmt{returnStmt, printReturnsStmt}, f.Body.List[len(f.Body.List)-1:]...),
			})
		}
	}
}

func main() {
	// 处理当前包
	if err := instrumentPackage("./instrument/math.go"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
