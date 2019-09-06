package httptest

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ---------------------------------------------------------------------------

type TransportComposer interface {
	Compose(base http.RoundTripper) http.RoundTripper
}

type Executor interface {
	Exec(ctx *Context, code string)
}

// ---------------------------------------------------------------------------

func mimeType(ct string) string {

	if ct == "form" {
		return "application/x-www-form-urlencoded"
	}
	if ct == "binary" {
		return "application/octet-stream"
	}
	if strings.Index(ct, "/") < 0 {
		return "application/" + ct
	}
	return ct
}

// ---------------------------------------------------------------------------

type Request struct {
	method   string
	url      string
	auth     TransportComposer
	ctx      *Context
	header   http.Header
	bodyType string
	body     string
}

func NewRequest(ctx *Context, method, url string) *Request {

	ctx.DeleteVar("resp")

	p := &Request{
		ctx: ctx,
		method: method,
		url: url,
		header: make(http.Header),
	}
	ctx.Log(" ====>", method, url)
	return p
}

func (p *Request) WithAuth(v interface{}) *Request {

	if v == nil {
		p.auth = nil
		return p
	}
	if name, ok := v.(string); ok {
		auth, ok := p.ctx.auths[name]
		if !ok {
			p.ctx.Fatal("WithAuth failed: auth not found -", name)
		}
		p.auth = auth
		return p
	}
	if auth, ok := v.(TransportComposer); ok {
		p.auth = auth
		return p
	}
	p.ctx.Fatal("WithAuth failed: invalid auth -", v)
	return p
}

func (p *Request) WithHeader(key string, values ...string) *Request {

	p.header[key] = values
	return p
}

func (p *Request) WithBody(bodyType, body string) *Request {

	p.bodyType = mimeType(bodyType)
	p.body = body
	return p
}

func (p *Request) WithBodyf(bodyType, format string, v ...interface{}) *Request {

	p.bodyType = mimeType(bodyType)
	p.body = fmt.Sprintf(format, v...)
	return p
}

func mergeHeader(to, from http.Header) {

	for k, v := range from {
		to[k] = v
	}
}

func (p *Request) send() (resp *http.Response, err error) {

	var body io.Reader
	if len(p.body) > 0 {
		body = strings.NewReader(p.body)
	}
	req, err := p.ctx.newRequest(p.method, p.url, body)
	if err != nil {
		p.ctx.Fatal("http.NewRequest failed:", p.method, p.url, p.body, err)
		return
	}

	mergeHeader(req.Header, p.ctx.DefaultHeader)

	if body != nil {
		if p.bodyType != "" {
			req.Header.Set("Content-Type", p.bodyType)
		}
		req.ContentLength = int64(len(p.body))
	}

	mergeHeader(req.Header, p.header)

	t := p.ctx.transport
	if p.auth != nil {
		t = p.auth.Compose(t)
	}

	c := &http.Client{Transport: t}
	return c.Do(req)
}

func (p *Request) Ret(code int) (resp *Response) {

	resp1, err := p.send()
	resp = newResponse(p, resp1, err)
	p.ctx.MatchVar("resp", map[string]interface{}{
		"body": resp.BodyObj,
		"header": resp.Header,
		"code": float64(resp.StatusCode),
	})
	return resp.matchCode(code)
}

// ---------------------------------------------------------------------------

type TestingT interface {
	Fatal(args ...interface{})
	Log(args ...interface{})
}

type NilTestingT struct {}

func (p NilTestingT) Fatal(args ...interface{}) {}
func (p NilTestingT) Log(args ...interface{}) {}

// ---------------------------------------------------------------------------

type Context struct {
	TestingT
	varsMgr
	hostsMgr
	transport          http.RoundTripper
	auths              map[string]TransportComposer
	DefaultHeader      http.Header
	MatchResponseError func(message string, req *Request, resp *Response)
}

func New(t TestingT) *Context {

	auths := make(map[string]TransportComposer)
	p := &Context{
		TestingT: t,
		auths: auths,
		transport: http.DefaultTransport,
		DefaultHeader: make(http.Header),
		MatchResponseError: matchRespError,
	}
	p.initHostsMgr()
	p.initVarsMgr()
	return p
}

func (p *Context) SetTransport(transport http.RoundTripper) {

	p.transport = transport
}

func (p *Context) SetAuth(name string, auth TransportComposer) {

	p.auths[name] = auth
}

func (p *Context) Exec(executor Executor, code string) *Context {

	executor.Exec(p, code)
	return p
}

func (p *Context) Request(method, url string) *Request {

	return NewRequest(p, method, url)
}

func (p *Context) Requestf(method, format string, v ...interface{}) *Request {

	url := fmt.Sprintf(format, v...)
	return NewRequest(p, method, url)
}

// ---------------------------------------------------------------------------

