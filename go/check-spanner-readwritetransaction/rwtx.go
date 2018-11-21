package rwtx

import (
	"fmt"
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"strings"
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
		(*ast.SelectorExpr)(nil),
	}

	inspect.Preorder(nodeFilter, func(node ast.Node) {
		switch node := node.(type) {
		case *ast.SelectorExpr:
			if node.Sel.Name != "ReadWriteTransaction" {
				return
			}

			xType, ok := pass.TypesInfo.Selections[node]
			if !ok {
				return
			}

			xPkg := xType.Obj().Pkg()
			if xPkg == nil {
				return
			}

			xPath := strings.TrimLeft(xPkg.Path(), pass.Pkg.Name()+"/vendor/")
			if xPath != "cloud.google.com/go/spanner" {
				return
			}

			obj, ok := pass.TypesInfo.Uses[node.Sel]
			if !ok {
				return
			}

			rwtxSig, ok := obj.Type().(*types.Signature)
			if !ok {
				return
			}
			rwtxArgs := rwtxSig.Params()
			if rwtxArgs.Len() != 2 {
				return
			}
			callbackFunc := rwtxArgs.At(1)
			callbackSig, ok := callbackFunc.Type().(*types.Signature)
			if !ok {
				return
			}
			callbackArgs := callbackSig.Params()
			if callbackArgs.Len() != 2 {
				return
			}
			txn := callbackArgs.At(1)



			fmt.Println(txn.Name())
		}
	})

	return nil, nil
}
