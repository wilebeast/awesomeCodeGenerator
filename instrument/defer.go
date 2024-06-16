package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"strings"
)

func instrumentFile(filePath string) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		fmt.Println("Error parsing file:", err)
		return
	}

	// 遍历抽象语法树,并在函数入口处插入 defer 语句
	instrumentFunctions(f)

	// 将修改后的文件写回磁盘
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	if err := printer.Fprint(file, fset, f); err != nil {
		fmt.Println("Error writing file:", err)
	}
}

func instrumentFunctions(f *ast.File) {
	ast.Inspect(f, func(node ast.Node) bool {
		if function, ok := node.(*ast.FuncDecl); ok {
			// 分析函数的输入参数
			inputParams := analyzeInputParams(function.Type.Params)

			// 分析函数的输出参数
			outputParams := analyzeOutputParams(function.Type.Results)

			// 在函数入口处插入 defer 语句
			insertDeferStatement(function, inputParams, outputParams)
		}
		return true
	})
}

func analyzeInputParams(fieldList *ast.FieldList) []string {
	var inputParams []string
	if fieldList != nil {
		for _, field := range fieldList.List {
			for _, name := range field.Names {
				inputParams = append(inputParams, name.Name)
			}
		}
	}
	return inputParams
}

func analyzeOutputParams(fieldList *ast.FieldList) []string {
	var outputParams []string
	if fieldList != nil {
		for i, field := range fieldList.List {
			if len(field.Names) == 0 {
				// 如果输出参数没有命名,则自动给它们命名
				paramName := fmt.Sprintf("X%d", i+1)
				field.Names = []*ast.Ident{ast.NewIdent(paramName)}
				outputParams = append(outputParams, paramName)
			} else {
				for _, name := range field.Names {
					outputParams = append(outputParams, name.Name)
				}
			}
		}
	}
	return outputParams
}

func formatArgInfos(params []string) string {
	var formattedArgs []string
	for _, param := range params {
		formattedArgs = append(formattedArgs, fmt.Sprintf("\"%s\":%s", param, param))
	}
	return fmt.Sprintf("map[string]interface{}{%s}", strings.Join(formattedArgs, ", "))
}

func insertDeferStatement(function *ast.FuncDecl, inputParams, outputParams []string) {
	// 生成 defer 语句的字符串
	deferStmt := "    ellen.Printf(" + "\"" + function.Name.Name + "\"," +
		formatArgInfos(inputParams) + "," + formatArgInfos(outputParams) + ")"

	// 在函数入口处插入 defer 语句
	function.Body.List = append([]ast.Stmt{&ast.DeferStmt{Call: &ast.CallExpr{
		Fun: &ast.FuncLit{
			Type: &ast.FuncType{},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ExprStmt{
						X: &ast.BasicLit{
							Kind:  token.STRING,
							Value: deferStmt,
						},
					},
				},
			},
		},
	}}}, function.Body.List...)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <file_path>")
		return
	}
	filePath := os.Args[1]
	instrumentFile(filePath)
}
