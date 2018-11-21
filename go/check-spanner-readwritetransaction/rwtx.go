package rwtx

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name: "spanner-rwtx",
	Doc:  "detect useless ReadWriteTransaction that can change into ReadTransaction",
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
		(*ast.FuncLit)(nil),
	}

	inspect.Preorder(nodeFilter, func(node ast.Node) {

		var funcType *ast.FuncType
		var funcBody *ast.BlockStmt

		switch node := node.(type) {
		case *ast.FuncDecl:
			funcType = node.Type
			funcBody = node.Body

		case *ast.FuncLit:
			funcType = node.Type
			funcBody = node.Body

		case *ast.CallExpr:

		default:
			return
		}

		for _, param := range funcType.Params.List {
			t, ok := pass.TypesInfo.Types[param.Type]
			if !ok {
				continue
			}
			if !isSpannerPtrReadWriteTransaction(pass, t.Type) {
				continue
			}

			for _, name := range param.Names {
				if !hasWriteTxCall(pass, node, name, funcBody) {
					pass.Reportf(node.Pos(), "this function never calls *spanner.ReadWriteTransaction's BufferWrite or Update method")
				}
			}
		}
	})

	return nil, nil
}

func isSpannerPtrReadWriteTransaction(pass *analysis.Pass, varDecl types.Type) bool {
	// *cloud.google.com/go/spanner.ReadWriteTransaction

	ptr, ok := varDecl.(*types.Pointer)
	if !ok {
		return false
	}
	named, ok := ptr.Elem().(*types.Named)
	if !ok {
		return false
	}

	if named.Obj().Name() != "ReadWriteTransaction" {
		return false
	}

	xPkg := named.Obj().Pkg()
	if xPkg == nil {
		return false
	}

	xPath := strings.TrimLeft(xPkg.Path(), pass.Pkg.Name()+"/vendor/")
	if xPath != "cloud.google.com/go/spanner" {
		return false
	}

	return true
}

func hasWriteTxCall(pass *analysis.Pass, node ast.Node, name *ast.Ident, funcBody *ast.BlockStmt) bool {
	v := &methodCallFindVisitor{
		// https://godoc.org/cloud.google.com/go/spanner#ReadWriteTransaction
		targetFunctionNames: []string{"BufferWrite", "Update"},
		node:                node,
		pass:                pass,
		name:                name,
	}
	ast.Walk(v, funcBody)
	return v.result
}

type methodCallFindVisitor struct {
	targetFunctionNames []string

	node ast.Node
	pass *analysis.Pass
	name *ast.Ident

	result bool
}

func (v *methodCallFindVisitor) Visit(node ast.Node) (w ast.Visitor) {
	if v.result {
		return nil
	}

	switch node := node.(type) {
	case *ast.CallExpr:
		switch fun := node.Fun.(type) {
		case *ast.SelectorExpr:
			found := false
			for _, targetFunctionName := range v.targetFunctionNames {
				if fun.Sel.Name == targetFunctionName {
					found = true
					break
				}
			}
			if !found {
				return v
			}

			ident, ok := fun.X.(*ast.Ident)
			if !ok {
				return v
			}
			if ident.Name != v.name.Name {
				// TODO 変数がshadowされたり別の名前に変えられたりしていた場合のことはとりあえず考えない
				return v
			}

			v.result = true
			return nil

		case *ast.Ident:
			// TODO identなどからFuncBodyを手に入れてチェックしたい…
			return nil
		}

	case *ast.FuncLit:
		// 関数内で関数が定義されても呼び出しされない限り中を見ない(見てもしょうがない)
		return nil
	}
	return v
}
