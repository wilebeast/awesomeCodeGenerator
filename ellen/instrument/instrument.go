package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"strconv"
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
			if function, ok := node.(*ast.FuncDecl); ok {
				for _, stmt := range function.Body.List {
					if deferStmt, ok := stmt.(*ast.DeferStmt); ok {
						if funcLit, ok := deferStmt.Call.Fun.(*ast.FuncLit); ok {
							if len(funcLit.Body.List) == 0 {
								continue
							}
							if exprStmt, ok := funcLit.Body.List[0].(*ast.ExprStmt); ok {
								if callExpr, ok := exprStmt.X.(*ast.CallExpr); ok {
									if selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
										if ident, ok := selectorExpr.X.(*ast.Ident); ok {
											// 找到包含 ellen.Printf 调用的 defer 语句
											if ident.Name == "ellen" && selectorExpr.Sel.Name == "Printf" {
												return false
											}
										}
									}
								}
							}
						}

					}
				}
			}

			// 分析函数的输入参数
			inputParams := analyzeInputParams(function.Type.Params)
			// 检查是否包含名为 "ctx" 的输入参数
			hasContextParam := false
			for _, param := range inputParams {
				if param == "ctx" {
					hasContextParam = true
					break
				}
			}
			// 如果包含 "ctx" 参数,则插入 defer 语句
			if hasContextParam {
				// 分析函数的输出参数
				outputParams := analyzeOutputParams(function.Type.Results)

				// 在函数入口处插入 defer 语句
				insertDeferStatement(function, inputParams, outputParams)

				ensurePackageImport(f, "code.byted.org/shark/ruleplatform/ellen")
			}
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

func ensurePackageImport(f *ast.File, pkgPath string) {
	// 检查是否已经导入了 pkgPath 包
	for _, imp := range f.Imports {
		if imp.Path.Value == strconv.Quote(pkgPath) {
			return
		}
	}

	// 如果没有导入 pkgPath 包,则添加一个导入语句
	f.Imports = append(f.Imports, &ast.ImportSpec{
		Path: &ast.BasicLit{
			Kind:  token.STRING,
			Value: strconv.Quote(pkgPath),
		},
	})

	// 在 f.Decls 中查找 token.IMPORT 类型的 ast.GenDecl 节点
	var importDecl *ast.GenDecl
	for _, decl := range f.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
			importDecl = genDecl
			break
		}
	}

	// 如果没有找到 token.IMPORT 类型的 ast.GenDecl 节点,则创建一个新的
	if importDecl == nil {
		importDecl = &ast.GenDecl{
			Tok: token.IMPORT,
		}
		f.Decls = append([]ast.Decl{importDecl}, f.Decls...)
	}

	// 创建新的导入语句并添加到 importDecl.Specs 中
	importSpec := &ast.ImportSpec{
		Path: &ast.BasicLit{
			Kind:  token.STRING,
			Value: strconv.Quote(pkgPath),
		},
	}
	importDecl.Specs = append(importDecl.Specs, importSpec)
}

func clearOutputParam(fieldList *ast.FieldList) {
	if fieldList != nil {
		for _, field := range fieldList.List {
			if len(field.Names) > 0 {
				// 遍历每个名字,如果能匹配为"X%d",则将field.Names设置为空
				for _, name := range field.Names {
					if strings.HasPrefix(name.Name, "X") && len(name.Name) > 1 {
						if _, err := strconv.Atoi(name.Name[1:]); err == nil {
							field.Names = nil
							break
						}
					}
				}
			}
		}
	}
}

func clearInstrumentFunctions(f *ast.File) {
	ast.Inspect(f, func(node ast.Node) bool {
		if function, ok := node.(*ast.FuncDecl); ok {
			for i, stmt := range function.Body.List {
				if deferStmt, ok := stmt.(*ast.DeferStmt); ok {
					if funcLit, ok := deferStmt.Call.Fun.(*ast.FuncLit); ok {
						if len(funcLit.Body.List) == 0 {
							continue
						}
						if exprStmt, ok := funcLit.Body.List[0].(*ast.ExprStmt); ok {
							if callExpr, ok := exprStmt.X.(*ast.CallExpr); ok {
								if selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
									if ident, ok := selectorExpr.X.(*ast.Ident); ok {
										// 找到包含 ellen.Printf 调用的 defer 语句
										if ident.Name == "ellen" && selectorExpr.Sel.Name == "Printf" {
											// 从函数体中移除该 defer 语句
											function.Body.List = append(function.Body.List[:i], function.Body.List[i+1:]...)
											// 清理自动命名参数
											clearOutputParam(function.Type.Results)
											break
										}
									}
								}
							}
						}
					}

				}
			}
		}
		return true
	})
}
func clearInstrumentedFile(filePath string) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		fmt.Println("Error parsing file:", err)
		return
	}

	// 遍历抽象语法树,并在函数入口处插入 defer 语句
	clearInstrumentFunctions(f)

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

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go <file_path> <clear_or_add>")
		return
	}
	filePath := os.Args[1]
	clearOrAdd := os.Args[2]

	if clearOrAdd == "clear" {
		clearInstrumentedFile(filePath)
	} else if clearOrAdd == "add" {
		instrumentFile(filePath)
	} else {
		fmt.Println("Invalid argument. Please use 'clear' or 'add'.")
	}
}
