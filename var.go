package httptest

import (
	"fmt"
	"io"
	"strings"

	"github.com/qiniu/dyn/jsonext"
	"github.com/qiniu/dyn/vars"
	"github.com/qiniu/x/log"

	. "github.com/qiniu/dyn/cmdarg"
)

const (
	Fmttype_Json    = vars.Fmttype_Json
	Fmttype_Form    = vars.Fmttype_Form
	Fmttype_Text    = vars.Fmttype_Text
	Fmttype_Jsonstr = vars.Fmttype_Jsonstr // 在json的字符串内
)

// ---------------------------------------------------------------------------

func PrettyPrintln(fprintln func(...interface{}) (int, error), values ...interface{}) {

	texts := make([]interface{}, len(values))
	for i, val := range values {
		if str, ok := val.(string); ok {
			texts[i] = str
			continue
		}
		text, err := jsonext.MarshalIndentToString(val, "", "  ")
		if err != nil {
			log.Warn("Fprintln: MarshalToString failed -", err, "val:", val)
		}
		texts[i] = text
	}
	fprintln(texts...)
}

func Fprintln(writer io.Writer, values ...interface{}) {

	fprintln := func(v ...interface{}) (int, error) {
		return fmt.Fprintln(writer, v...)
	}
	PrettyPrintln(fprintln, values...)
}

func Println(values ...interface{}) {

	PrettyPrintln(fmt.Println, values...)
}

// ---------------------------------------------------------------------------

type varsMgr struct {
	*vars.Context
}

func (p *varsMgr) initVarsMgr() {

	p.Context = vars.New()
}

// ---------------------------------------------------------------------------

func (p *Context) GetVar(key string) Var {

	v1, ok := p.varsMgr.GetVar(key)
	return Var{v1, ok}
}

func (p *Context) Requestv(method, urlWithVar string) *Request {

	url, err := p.SubstText(urlWithVar, Fmttype_Form)
	if err != nil {
		p.Fatal("invalid request url:", err)
	}
	return NewRequest(p, method, url)
}

// ---------------------------------------------------------------------------

func (p *Request) WithHeaderv(key string, valuesVar ...string) (resp *Request) {

	if len(valuesVar) == 1 {
		valVar := valuesVar[0]
		if strings.HasPrefix(valVar, "$(") && strings.HasSuffix(valVar, ")") {
			valKey := valVar[2 : len(valVar)-1]
			if val, ok := p.ctx.varsMgr.GetVar(valKey); ok {
				if varr, ok := val.([]string); ok {
					return p.WithHeader(key, varr...)
				}
			}
		}
	}

	values := make([]string, len(valuesVar))
	for i, valVar := range valuesVar {
		val, err := p.ctx.SubstText(valVar, Fmttype_Text)
		if err != nil {
			p.ctx.Fatal("invalid request header:", err, "key:", key, "value:", valVar)
			return p
		}
		values[i] = val
	}

	return p.WithHeader(key, values...)
}

func (p *Request) WithBodyv(bodyType, bodyWithVar string) *Request {

	var ft int

	bodyType = mimeType(bodyType)
	switch bodyType {
	case "application/json":
		ft = Fmttype_Json
	case "application/x-www-form-urlencoded":
		ft = Fmttype_Form
	default:
		ft = Fmttype_Text
	}
	body, err := p.ctx.SubstText(bodyWithVar, ft)
	if err != nil {
		p.ctx.Fatal("invalid request body:", err)
		return p
	}
	return p.WithBody(bodyType, body)
}

// ---------------------------------------------------------------------------

func (p *Response) WithHeaderv(key string, valuesVar ...string) (resp *Response) {

	values := make([]interface{}, len(valuesVar))
	for i, valVar := range valuesVar {
		val, err := UnmarshalText(valVar)
		if err != nil {
			p.matchRespError("unmarshal failed: " + err.Error() + " - text: " + valVar)
			return p
		}
		values[i] = val
	}

	err := p.ctx.Match(values, p.Header[key])
	if err != nil {
		p.matchRespError("match header failed: " + err.Error())
	}
	return p
}

func (p *Response) WithBodyv(bodyType, bodyVar string) (resp *Response) {

	bodyExpected, err := Unmarshal(bodyVar)
	if err != nil {
		p.matchRespError("unmarshal failed: " + err.Error() + " - json text: " + bodyVar)
		return p
	}

	err = p.ctx.Match(bodyExpected, p.BodyObj)
	if err != nil {
		p.matchRespError("match response body failed: " + err.Error())
	}
	return p
}

// ---------------------------------------------------------------------------
