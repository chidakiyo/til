package main

import (
	"github.com/vvakame/til/go/check-spanner-readwritetransaction"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(rwtx.Analyzer)
}
