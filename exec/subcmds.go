package exec

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/qiniu/dyn/cmdarg"
	"github.com/qiniu/httptest"
)

// ---------------------------------------------------------------------------

type base64Args struct {
	StdEncoding bool   `flag:"std - use standard base64 encoding. default is urlsafe base64 encoding."`
	Fdecode     bool   `flag:"d - to decode data. default is to encode data."`
	Data        string `arg:"data"`
}

func (p *subContext) Eval_base64(ctx *httptest.Context, args *base64Args) (string, error) {

	encoding := base64.URLEncoding
	if args.StdEncoding {
		encoding = base64.StdEncoding
	}
	if args.Fdecode {
		b, err := encoding.DecodeString(args.Data)
		if err != nil {
			return "", err
		}
		return string(b), nil
	} else {
		return encoding.EncodeToString([]byte(args.Data)), nil
	}
}

// ---------------------------------------------------------------------------

type envArgs struct {
	VarName string `arg:"var-name"`
}

func (p *subContext) Eval_env(ctx *httptest.Context, args *envArgs) (string, error) {

	v := os.Getenv(args.VarName)
	if v != "" {
		return v, nil
	}
	return "", fmt.Errorf("env `%s` not found", args.VarName)
}

// ---------------------------------------------------------------------------

type decodeArgs struct {
	Text string `arg:"text"`
}

func (p *subContext) Eval_decode(ctx *httptest.Context, args *decodeArgs) (interface{}, error) {

	return cmdarg.Unmarshal(args.Text)
}

// ---------------------------------------------------------------------------

type envdecodeArgs struct {
	VarName string `arg:"var-name"`
}

func (p *subContext) Eval_envdecode(ctx *httptest.Context, args *envdecodeArgs) (interface{}, error) {

	v := os.Getenv(args.VarName)
	if v != "" {
		return cmdarg.Unmarshal(v)
	}
	return nil, fmt.Errorf("env `%s` not found", args.VarName)
}

// ---------------------------------------------------------------------------
