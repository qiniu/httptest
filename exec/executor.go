package exec

import (
	"reflect"
	"strings"

	"github.com/qiniu/dyn/flag"
	"github.com/qiniu/httptest"
	"github.com/qiniu/x/cmdline"
)

// ---------------------------------------------------------------------------

type IContext interface {
	GetRawCmd() string
}

type IExternalContext interface {
	FindCmd(ctx IContext, cmd string) reflect.Value
}

var (
	External    IExternalContext
	ExternalSub IExternalContext
)

// ---------------------------------------------------------------------------

type Context struct {
	rawCmd  string
	current interface{}
	autoVarMgr
}

func New() *Context {

	return &Context{}
}

func (p *Context) Exec(ctx *httptest.Context, code string) {

	sctx := &subContext{
		ctx:    ctx,
		parent: p,
	}
	sctx.parser = cmdline.NewParser()
	sctx.parser.ExecSub = sctx.execSubCmd

retry:
	code, err := p.parseAndExec(ctx, sctx, code)
	if err == nil {
		goto retry
	}
}

func (p *Context) GetRawCmd() string {

	return p.rawCmd
}

func (p *Context) findCmd(cmd string) (method reflect.Value) {

	v := reflect.ValueOf(p)
	method = v.MethodByName("Cmd_" + cmd)
	if method.IsValid() {
		return
	}

	if External == nil {
		return
	}
	return External.FindCmd(p, cmd)
}

func (p *Context) parseAndExec(
	ctx *httptest.Context, sctx *subContext, code string) (codeNext string, err error) {

	baseFrame := p.enterFrame()
	defer p.leaveFrame(ctx, baseFrame)

	cmd, codeNext, err := sctx.parser.ParseCode(code)
	if err != nil && err != cmdline.EOF {
		ctx.Fatal(err)
		return
	}
	if len(cmd) > 0 {
		//
		// p.Cmd_xxx(ctx *httptest.Context, cmd []string)
		method := p.findCmd(cmd[0])
		if !method.IsValid() {
			ctx.Fatal("command not found:", cmd[0])
			return
		}
		cmdLen := len(code) - len(codeNext)
		p.rawCmd = strings.Trim(code[:cmdLen], " \t\r\n")
		ctx.Log("====>", p.rawCmd)
		_, err = runCmd(ctx, method, cmd)
		if err != nil {
			ctx.Fatal(cmd, "-", err)
			return
		}
	}
	return
}

func runCmd(ctx *httptest.Context, method reflect.Value, cmd []string) (out []reflect.Value, err error) {

	return flag.ExecMethod(ctx.Context, method, reflect.ValueOf(ctx), cmd)
}

// ---------------------------------------------------------------------------
