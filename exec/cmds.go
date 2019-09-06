package exec

import (
	"fmt"
	"strings"

	"github.com/qiniu/httptest"
)

// ---------------------------------------------------------------------------

type hostArgs struct {
	Host   string `arg:"host - eg. api.qiniu.com"`
	Portal string `arg:"portal - eg. <ip>:<port>"`
}

func (p *Context) Cmd_host(ctx *httptest.Context, args *hostArgs) {

	ctx.SetHost(args.Host, args.Portal)
}

// ---------------------------------------------------------------------------

type authArgs struct {
	AuthInfo      interface{} `arg:"auth-information"`
	AuthInterface interface{} `arg:"auth-interface,opt"`
}

func (p *Context) Cmd_auth(ctx *httptest.Context, args *authArgs) {

	if args.AuthInterface == nil {
		if req, ok := p.current.(*httptest.Request); ok {
			req.WithAuth(args.AuthInfo)
		} else {
			ctx.Fatal("incorrect context to call `auth <auth-information>`")
		}
	} else {
		if name, ok := args.AuthInfo.(string); ok {
			if auth, ok := args.AuthInterface.(httptest.TransportComposer); ok {
				ctx.SetAuth(name, auth)
				return
			}
		}
		ctx.Fatal("usage: auth <auth-name> <auth-interface>")
	}
}

// ---------------------------------------------------------------------------

type printlnArgs struct {
	Values []interface{} `arg:value`
}

func (p *Context) Cmd_println(ctx *httptest.Context, args *printlnArgs) {

	p.Cmd_echo(ctx, args)
}

func (p *Context) Cmd_echo(ctx *httptest.Context, args *printlnArgs) {

	fprintln := func(v ...interface{}) (int, error) {
		ctx.Log(v...)
		return fmt.Println(v...)
	}
	httptest.PrettyPrintln(fprintln, args.Values...)
}

// ---------------------------------------------------------------------------

type req1Args struct {
	Url string `arg:url`
}

type reqArgs struct {
	Method string `arg:method`
	Url    string `arg:url`
}

func (p *Context) Cmd_req(ctx *httptest.Context, args *reqArgs) {

	p.current = ctx.Request(strings.ToUpper(args.Method), args.Url)
}

func (p *Context) Cmd_post(ctx *httptest.Context, args *req1Args) {

	p.current = ctx.Request("POST", args.Url)
}

func (p *Context) Cmd_get(ctx *httptest.Context, args *req1Args) {

	p.current = ctx.Request("GET", args.Url)
}

func (p *Context) Cmd_delete(ctx *httptest.Context, args *req1Args) {

	p.current = ctx.Request("DELETE", args.Url)
}

func (p *Context) Cmd_put(ctx *httptest.Context, args *req1Args) {

	p.current = ctx.Request("PUT", args.Url)
}

// ---------------------------------------------------------------------------

type headerArgs struct {
	Key    string   `arg:"key"`
	Values []string `arg:"value,keep"`
}

func (p *Context) Cmd_header(ctx *httptest.Context, args *headerArgs) {

	if req, ok := p.current.(*httptest.Request); ok {
		req.WithHeaderv(args.Key, args.Values...)
	} else if resp, ok := p.current.(*httptest.Response); ok {
		resp.WithHeaderv(args.Key, args.Values...)
	} else {
		ctx.Fatal("incorrect context to call `header <key> <value1> <value2> ...`")
	}
}

// ---------------------------------------------------------------------------

type body1Args struct {
	Body string `arg:"body,keep"` // keep: 保留 $(var) 不要自动展开
}

type bodyArgs struct {
	BodyType string `arg:"body-type - eg. json, form, application/json, application/text, etc"`
	Body     string `arg:"body,keep"` // keep: 保留 $(var) 不要自动展开
	Keep     bool   `flag:"pure-text - keep $(var) as pure text"`
}

func (p *Context) Cmd_body(ctx *httptest.Context, args *bodyArgs) {

	p.withBodyv(ctx, args.BodyType, args.Body, args.Keep)
}

func (p *Context) Cmd_json(ctx *httptest.Context, args *body1Args) {

	p.withBodyv(ctx, "json", args.Body, false)
}

func (p *Context) Cmd_bson(ctx *httptest.Context, args *body1Args) {

	p.withBodyv(ctx, "bson", args.Body, false)
}

func (p *Context) Cmd_form(ctx *httptest.Context, args *body1Args) {

	if req, ok := p.current.(*httptest.Request); ok {
		p.current = req.WithBodyv("form", args.Body)
	} else {
		ctx.Fatal("incorrect context to call `form <form-body>`")
	}
}

func (p *Context) Cmd_binary(ctx *httptest.Context, args *body1Args) {

	if req, ok := p.current.(*httptest.Request); ok {
		p.current = req.WithBodyv("binary", args.Body)
	} else {
		ctx.Fatal("incorrect context to call `binary <binary-body>`")
	}
}

func (p *Context) withBodyv(ctx *httptest.Context, bodyType, body string, keep bool) {

	if req, ok := p.current.(*httptest.Request); ok {
		if keep {
			req.WithBody(bodyType, body)
		} else {
			req.WithBodyv(bodyType, body)
		}
	} else if resp, ok := p.current.(*httptest.Response); ok {
		if keep {
			resp.WithBody(bodyType, body)
		} else {
			resp.WithBodyv(bodyType, body)
		}
	} else {
		ctx.Fatal("incorrect context to call:", p.rawCmd)
	}
}

// ---------------------------------------------------------------------------

type retArgs struct {
	Code int `arg:"code,opt"` // opt: 可选参数
}

func (p *Context) Cmd_ret(ctx *httptest.Context, args *retArgs) {

	if req, ok := p.current.(*httptest.Request); ok {
		p.current = req.Ret(args.Code)
	} else {
		ctx.Fatal("incorrect context to call `ret <code>`")
	}
}

// ---------------------------------------------------------------------------

type clearArgs struct {
	VarNames []string `arg:var-name`
}

func (p *Context) Cmd_clear(ctx *httptest.Context, args *clearArgs) {

	for _, varName := range args.VarNames {
		ctx.DeleteVar(varName)
	}
}

// ---------------------------------------------------------------------------

type matchArgs struct {
	Expected interface{} `arg:"expected-object,keep"` // keep: 不要做 Subst
	Source   interface{} `arg:"source-object"`
}

func (p *Context) Cmd_match(ctx *httptest.Context, args *matchArgs) {

	err := ctx.Match(args.Expected, args.Source)
	if err != nil {
		expected := substObject(ctx, args.Expected, httptest.Fmttype_Text)
		ctx.Fatal("match failed:", err, "-", expected, args.Source)
	}
}

func substObject(ctx *httptest.Context, obj interface{}, ft int) interface{} {

	if obj2, err := ctx.Subst(obj, ft); err == nil {
		return obj2
	}
	return obj
}

// ---------------------------------------------------------------------------

type letArgs struct {
	Expected interface{} `arg:"var,keep"` // keep: 不要做 Subst
	Source   interface{} `arg:"value"`
}

func (p *Context) Cmd_let(ctx *httptest.Context, args *letArgs) {

	err := ctx.Let(args.Expected, args.Source)
	if err != nil {
		ctx.Fatal("let failed:", err, "-", args.Expected, "value:", args.Source)
	}
}

// ---------------------------------------------------------------------------

type equalArgs struct {
	Object1 interface{} `arg:"object1"`
	Object2 interface{} `arg:"object2"`
}

func (p *Context) Cmd_equal(ctx *httptest.Context, args *equalArgs) {

	if !httptest.Equal(args.Object1, args.Object2) {
		ctx.Fatal("equal test failed:", p.rawCmd, "- objects:", args.Object1, args.Object2)
	}
}

func (p *Context) Cmd_equalSet(ctx *httptest.Context, args *equalArgs) {

	if !httptest.EqualSet(args.Object1, args.Object2) {
		ctx.Fatal("equalSet test failed:", p.rawCmd, "- objects:", args.Object1, args.Object2)
	}
}

// ---------------------------------------------------------------------------
