package rwtx_test

import (
	"github.com/vvakame/til/go/check-spanner-readwritetransaction"
	"golang.org/x/tools/go/analysis/analysistest"
	"testing"
)

func Test(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, rwtx.Analyzer, "a")
}
