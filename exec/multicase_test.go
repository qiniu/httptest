package exec

import (
	"reflect"
	"testing"
)

// ---------------------------------------------------------------------------

type caseParseCases struct {
	code  string
	cases Cases
	err   error
}

func TestParseCases(t *testing.T) {

	tests := []caseParseCases{
		{
			code: `
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
	`,
			cases: Cases{
				SetUp: `#代码片段1
	...

	`,
				TearDown: `#代码片段4
	...
	`,
				Items: []Case{
					{
						Name: "testCase1",
						Code: `#代码片段2
	...

	`,
					},
					{
						Name: "testCase2",
						Code: `#代码片段3
	...

	`,
					},
				},
			},
		},
	}
	for _, c := range tests {
		cases, err := ParseCases(c.code)
		if !reflect.DeepEqual(cases, c.cases) || err != c.err {
			t.Fatal("ParseCases failed:", cases, err)
		}
	}
}

// ---------------------------------------------------------------------------

