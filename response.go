package httptest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"

	"gopkg.in/mgo.v2/bson"
)

// ---------------------------------------------------------------------------

func matchJsonValue(expected, real interface{}) (string, bool) {

	if vexpected, ok := expected.(map[string]interface{}); ok {
		if vreal, ok2 := real.(map[string]interface{}); ok2 {
			for k, v := range vexpected {
				v3, ok3 := vreal[k]
				if !ok3 || !reflect.DeepEqual(v, v3) {
					return "unmatched json response body - field " + k, false
				}
			}
			return "", true
		}
	} else if reflect.DeepEqual(expected, real) {
		return "", true
	}
	return "unmatched json response body", false
}

func matchStream(a string, b []byte) bool {

	if len(b) != len(a) {
		return false
	}
	for i, c := range b {
		if a[i] != c {
			return false
		}
	}
	return true
}

func decode(in io.Reader) (v interface{}, err error) {

	dec := json.NewDecoder(in)
	err = dec.Decode(&v)
	return
}

// ---------------------------------------------------------------------------

type Response struct {
	ctx        *Context
	req        *Request
	Header     http.Header
	BodyType   string
	RawBody    []byte
	BodyObj    interface{}
	Err        error
	StatusCode int
}

func newResponse(req *Request, resp *http.Response, err error) (p *Response) {

	p = &Response{
		req: req,
		ctx: req.ctx,
		Err: err,
	}
	if err != nil {
		p.ctx.Fatal("http request failed:", err)
		return
	}
	p.StatusCode = resp.StatusCode
	p.Header = resp.Header
	defer resp.Body.Close()

	p.RawBody, p.Err = ioutil.ReadAll(resp.Body)
	if p.Err != nil {
		p.ctx.Fatal("read response body failed:", p.Err)
		return
	}
	p.BodyType = p.Header.Get("Content-Type")
	if len(p.RawBody) > 0 {
		switch p.BodyType {
		case "application/json":
			p.Err = json.Unmarshal(p.RawBody, &p.BodyObj)
			if p.Err != nil {
				p.ctx.Fatal("unmarshal response body failed:", p.Err)
			}
		case "application/bson":
			p.Err = bson.Unmarshal(p.RawBody, &p.BodyObj)
			if p.Err != nil {
				p.ctx.Fatal("unmarshal response body failed:", p.Err)
			}
			b, _ := json.Marshal(p.BodyObj)
			p.BodyObj = nil
			p.Err = json.Unmarshal(b, &p.BodyObj)
		default:
			p.BodyObj = string(p.RawBody)
		}
	}
	return
}

func (p *Response) matchCode(code int) (resp *Response) {

	resp = p
	if code != 0 && code != resp.StatusCode {
		p.matchRespError("unmatched status code")
	}
	return
}

func (p *Response) WithHeader(k string, v ...string) (resp *Response) {

	resp = p
	if !matchHeader(p.Header, k, v) {
		p.matchRespError("unmatched response header: " + k)
	}
	return
}

func matchHeader(header http.Header, k string, v []string) bool {

	if realv, ok := header[k]; ok {
		if reflect.DeepEqual(v, realv) {
			return true
		}
	}
	return false
}

func (p *Response) WithBodyf(bodyType, format string, v ...interface{}) (resp *Response) {

	return p.WithBody(bodyType, fmt.Sprintf(format, v...))
}

func (p *Response) WithBody(bodyType, body string) (resp *Response) {

	resp = p
	if message, ok := matchBody(mimeType(bodyType), body, p); !ok {
		p.matchRespError(message)
	}
	return
}

func matchBody(expectedBodyType, expectedBody string, resp *Response) (string, bool) {

	if expectedBodyType != "" {
		if expectedBodyType != resp.BodyType {
			return "unmatched Content-Type", false
		}
	}

	if expectedBodyType == "application/json" && expectedBody != "" {
		v, err := decode(strings.NewReader(expectedBody))
		if err != nil {
			return "expected body isn't a valid json: " + err.Error(), false
		}
		return matchJsonValue(v, resp.BodyObj)
	}

	if !matchStream(expectedBody, resp.RawBody) {
		return "unmatched response body", false
	}
	return "", true
}

func (p *Response) matchRespError(message string) {

	p.Err = errors.New(message)
	p.ctx.MatchResponseError(message, p.req, p)
}

func matchRespError(message string, p *Request, resp *Response) {

	p.ctx.Fatal(
		message, "- req:", *p, "- resp:", resp.StatusCode, resp.Header, string(resp.RawBody))
}

// ---------------------------------------------------------------------------

func (p *Response) GetBody(v interface{}) (resp *Response) {

	resp = p
	switch ct := p.Header.Get("Content-Type"); ct {
	case "application/json":
		err := json.Unmarshal(p.RawBody, v)
		if err != nil {
			p.ctx.Fatal("GetBody failed:", err)
		}
	default:
		p.ctx.Fatal("GetBody failed: unsupported Content-Type", ct)
	}
	return
}

// ---------------------------------------------------------------------------
