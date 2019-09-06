package exec

import (
	"strconv"

	"github.com/qiniu/httptest"
)

// ---------------------------------------------------------------------------

type autoVarMgr struct {
	varNo int64
}

func (p *autoVarMgr) enterFrame() (baseFrame int64) {

	return p.varNo
}

func (p *autoVarMgr) substObject(ctx *httptest.Context, v interface{}) string {

	p.varNo++
	varName := getAutoVarName(p.varNo)
	err := ctx.MatchVar(varName, v)
	if err != nil {
		ctx.Fatal("create auto variable failed:", err)
	}
	return "$(" + varName + ")"
}

func (p *autoVarMgr) leaveFrame(ctx *httptest.Context, baseFrame int64) {

	for varNo := baseFrame + 1; varNo <= p.varNo; varNo++ {
		varName := getAutoVarName(varNo)
		ctx.DeleteVar(varName)
	}
}

func getAutoVarName(varNo int64) string {

	return "--auto-var-" + strconv.FormatInt(varNo, 10)
}

// ---------------------------------------------------------------------------
