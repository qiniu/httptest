package exec

import (
	"errors"
	"reflect"

	"github.com/qiniu/httptest"
	"github.com/qiniu/x/cmdline"
)

var (
	ErrSubCmdNotFound = errors.New("sub command not found")
)

// ---------------------------------------------------------------------------

type subContext struct {
	parent *Context
	parser *cmdline.Parser
	ctx    *httptest.Context
	rawCmd string
}

func (p *subContext) GetRawCmd() string {

	return p.rawCmd
}

func (p *subContext) findCmd(cmd string) (method reflect.Value) {

	v := reflect.ValueOf(p)
	method = v.MethodByName("Eval_" + cmd)
	if method.IsValid() {
		return
	}

	if ExternalSub == nil {
		return
	}
	return ExternalSub.FindCmd(p, cmd)
}

func (p *subContext) execSubCmd(code string) (val string, err error) {

	cmd, err := p.parser.ParseCmd(code)
	if err != nil {
		return
	}

	//
	// p.Eval_xxx(ctx *httptest.Context, cmd []string) (interface{}, error)
	method := p.findCmd(cmd[0])
	if !method.IsValid() {
		return "", ErrSubCmdNotFound
	}

	p.rawCmd = code
	out, err := runCmd(p.ctx, method, cmd)
	if err != nil {
		return
	}
	if len(out) != 2 {
		return "", ErrSubCmdNotFound
	}
	if !out[1].IsNil() {
		return "", out[1].Interface().(error)
	}

	return p.parent.substObject(p.ctx, out[0].Interface()), nil
}

// ---------------------------------------------------------------------------
