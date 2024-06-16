package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <source_file>")
		return
	}

	sourceFile := os.Args[1]

	fset := token.NewFileSet()

	// 解析源文件
	f, err := parser.ParseFile(fset, sourceFile, nil, parser.ParseComments)
	if err != nil {
		fmt.Println("Failed to parse file:", err)
		return
	}

	// 遍历文件中的函数声明
	for _, decl := range f.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			// 生成新的函数名
			newFuncName := fmt.Sprintf("%sWrapper", fn.Name.Name)

			// 创建新的函数声明
			newFunc := &ast.FuncDecl{
				Name: ast.NewIdent(newFuncName),
				Type: fn.Type,
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.ExprStmt{
							X: &ast.CallExpr{
								Fun: ast.NewIdent("fmt.Println"),
								Args: []ast.Expr{
									&ast.BasicLit{
										Kind:  token.STRING,
										Value: fmt.Sprintf(`"Input: %+v"`, fn.Type),
									},
								},
							},
						},
						fn.Body,
						&ast.ExprStmt{
							X: &ast.CallExpr{
								Fun: ast.NewIdent("fmt.Println"),
								Args: []ast.Expr{
									&ast.BasicLit{
										Kind:  token.STRING,
										Value: `"Output: "`,
									},
									&ast.BasicLit{
										Kind:  token.STRING,
										Value: "result",
									},
								},
							},
						},
					},
				},
			}

			// 将新函数插入到文件中
			f.Decls = append(f.Decls, newFunc)
		}
	}

	// 打印修改后的代码
	printer.Fprint(os.Stdout, fset, f)
}
