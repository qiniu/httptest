package exec

import (
	"errors"

	"github.com/qiniu/httptest"
	"github.com/qiniu/x/cmdline"

	. "github.com/qiniu/x/ctype"
)

var (
	ErrUnexpected           = errors.New("unexpected")
	ErrSyntaxError_Case     = errors.New("syntax error: please use `case <testCaseName>`")
	ErrSyntaxError_TearDown = errors.New("syntax error: please use `tearDown`")
)

// ---------------------------------------------------------------------------

/*
一般测试案例框架都有选择性执行某个案例、多个案例共享 setUp、tearDown 这样的启动和终止代码。我们也可以考虑支持。设想如下：

	#代码片段1
	...

	case testCase1
	#代码片段2
	...

	case testCase2
	#代码片段3
	...

	tearDown
	#代码片段4
	...

这段代码里面，“代码片段1” 将被认为是 setUp 代码，“代码片段4” 是 tearDown 代码，所有 testCase 开始前都会执行一遍“代码片段1”，退出前执行一遍“代码片段4”。每个 case 不用写 end 语句，遇到下一个 case 或者遇到 tearDown 就代表该 case 结束。
*/
type Case struct {
	Name string
	Code string
}

type Cases struct {
	SetUp    string
	TearDown string
	Items    []Case
}

func (p *Cases) Exec(ctx *httptest.Context, code string) {

	ectx := New()
	ectx.Exec(ctx, p.SetUp)
	if code != "" {
		ectx.Exec(ctx, code)
	}
	ectx.Exec(ctx, p.TearDown)
}

// ---------------------------------------------------------------------------

func ExecCases(t httptest.TestingT, code string) (err error) {

	cases, err := ParseCases(code)
	if err != nil {
		return
	}

	if len(cases.Items) == 0 {
		ctx := httptest.New(t)
		cases.Exec(ctx, "")
		return nil
	}

	for _, c := range cases.Items {
		ctx := httptest.New(t)
		ctx.Log("==========", c.Name, "===========")
		cases.Exec(ctx, c.Code)
	}
	return
}

// ---------------------------------------------------------------------------

const (
	endOfLine    = EOL | SEMICOLON // [\r\n;]
	symbols      = CSYMBOL_NEXT_CHAR
	blanks       = SPACE_BAR | TAB
	blankAndEOLs = SPACE_BAR | TAB | endOfLine
)

func ParseCases(code string) (cases Cases, err error) {

	seg, code, n := parseSeg(code)
	cases.SetUp = seg

	for code != "" {
		switch code[:n] {
		case "case":
			caseName, code2, err2 := parseCase(code[n:])
			if err2 != nil {
				err = err2
				return
			}
			seg, code, n = parseSeg(code2)
			cases.Items = append(cases.Items, Case{caseName, seg})
		case "tearDown":
			code2, err2 := parseTearDown(code[n:])
			if err2 != nil {
				err = err2
				return
			}
			seg, code, n = parseSeg(code2)
			cases.TearDown = seg
		default:
			err = ErrUnexpected
			return
		}
	}
	return
}

func parseSeg(code string) (seg string, codeNext string, n int) {

	code = cmdline.Skip(code, blankAndEOLs)
	codeNext = code
	for {
		n = cmdline.Find(codeNext, blankAndEOLs)
		switch codeNext[:n] {
		case "case", "tearDown", "":
			seg = code[:len(code)-len(codeNext)]
			return
		default:
			k := cmdline.Find(codeNext, endOfLine)
			codeNext = cmdline.Skip(codeNext[k:], blankAndEOLs)
		}
	}
}

func parseCase(code string) (caseName string, codeNext string, err error) {

	code = cmdline.Skip(code, blanks)
	symbolNext := cmdline.Skip(code, symbols)
	n := len(code) - len(symbolNext)
	if n > 0 {
		code2 := cmdline.Skip(symbolNext, blanks)
		if isEOL(code2) {
			return code[:n], code2, nil
		}
	}
	err = ErrSyntaxError_Case
	return
}

func parseTearDown(code string) (codeNext string, err error) {

	code = cmdline.Skip(code, blanks)
	if isEOL(code) {
		return code, nil
	}
	err = ErrSyntaxError_TearDown
	return
}

func isEOL(str string) bool {

	for _, c := range str {
		return Is(endOfLine, c)
	}
	return true
}

// ---------------------------------------------------------------------------
