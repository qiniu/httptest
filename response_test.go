package httptest

import (
	"encoding/json"
	"testing"
	"net/http"
)

// ---------------------------------------------------------------------------

type caseMatchHeader struct {
	expected http.Header
	real     http.Header
	message  string
	ok       bool
}

func matchHeaders(expected, real http.Header) (message string, ok bool) {

	for k, v := range expected {
		if !matchHeader(real, k, v) {
			return "unmatched response header: " + k, false
		}
	}
	return "", true
}

func TestMatchHeader(t *testing.T) {

	cases := []caseMatchHeader{
		{
			expected: http.Header{"a": {"a1"}, "b": {"b1", "b2"}},
			real: http.Header{"c": {"c1"}, "a": {"a1"}, "b": {"b1", "b2"}, "d": {"d1"}},
			ok: true,
		},
		{
			expected: http.Header{"a": {"a1"}, "b": {"b1", "b2"}},
			real: http.Header{"c": {"c1"}, "a": {"a1"}, "b": {"b1"}, "d": {"d1"}},
			message: "unmatched response header: b",
		},
		{
			expected: http.Header{"a": {"a1"}, "b": {"b1", "b2"}},
			real: http.Header{"c": {"c1"}, "a": {"a1"}, "d": {"d1"}},
			message: "unmatched response header: b",
		},
		{
			expected: http.Header{},
			real: http.Header{"c": {"c1"}, "a": {"a1"}, "d": {"d1"}},
			ok: true,
		},
	}
	for _, c := range cases {
		message, ok := matchHeaders(c.expected, c.real)
		if message != c.message || ok != c.ok {
			t.Fatal("matchHeader failed:", c, message, ok)
		}
	}
}

// ---------------------------------------------------------------------------

type caseMatchBody struct {
	expBody      string
	expBodyType  string
	respBody     string
	respBodyType string
	message      string
	ok           bool
}

func TestMatchBody(t *testing.T) {

	cases := []caseMatchBody{
		{
			expBody: ``,
			expBodyType: "application/json",
			respBody: ``,
			respBodyType: "application/json",
			ok: true,
		},
		{
			expBody: `abc`,
			expBodyType: "",
			respBody: `abc`,
			respBodyType: "application/json",
			ok: true,
		},
		{
			expBody: `abc`,
			expBodyType: "application/text",
			respBody: `abc`,
			respBodyType: "application/text",
			ok: true,
		},
		{
			expBody: `abc`,
			expBodyType: "application/text",
			respBody: `abc`,
			respBodyType: "application/json",
			message: "unmatched Content-Type",
		},
		{
			expBody: `{}`,
			expBodyType: "application/json",
			respBody: `{    }`,
			respBodyType: "application/json",
			ok: true,
		},
		{
			expBody: `{}`,
			expBodyType: "application/json",
			respBody: `{"aaa": "123"}`,
			respBodyType: "application/json",
			ok: true,
		},
		{
			expBody: `{"a": "a1", "b": ["b1", "b2"]}`,
			expBodyType: "application/json",
			respBody: `{"c": 1, "a": "a1", "b": ["b1", "b2"], "d": 2.0}`,
			respBodyType: "application/json",
			ok: true,
		},
		{
			expBody: `{"a": "a1", "b": ["b1", "b2"]}`,
			expBodyType: "application/json",
			respBody: `{"c": 1, "a": "a1", "d": 2.0}`,
			respBodyType: "application/json",
			message: "unmatched json response body - field b",
		},
		{
			expBody: `{"a": "a1", "b": ["b1", "b2"]}`,
			expBodyType: "application/json",
			respBody: `{"c": 1, "a": "a1", "b": [1], "d": 2.0}`,
			respBodyType: "application/json",
			message: "unmatched json response body - field b",
		},
	}
	for _, c := range cases {
		resp := &Response{
			RawBody: []byte(c.respBody),
			Header: make(http.Header),
		}
		if c.respBodyType != "" {
			resp.BodyType = c.respBodyType
			resp.Header["Content-Type"] = []string{c.respBodyType}
		}
		if len(resp.RawBody) > 0 {
			if resp.BodyType == "application/json" {
				resp.Err = json.Unmarshal(resp.RawBody, &resp.BodyObj)
			}
		}
		message, ok := matchBody(c.expBodyType, c.expBody, resp)
		if message != c.message || ok != c.ok {
			t.Fatal("matchBody failed:", c, message, ok)
		}
	}
}

// ---------------------------------------------------------------------------

