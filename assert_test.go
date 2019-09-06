package httptest

import (
	"testing"
)

// ---------------------------------------------------------------------------

func TestVar(t *testing.T) {

	if !(Var{1, true}.EqualObject("1")) {
		t.Fatal("EqualObject failed")
	}

	if !(Var{[]float64{1, 2}, true}.EqualSet("[2, 1]")) {
		t.Fatal("EqualSet failed")
	}

	if (Var{[]float64{1, 2, 3}, true}.EqualSet("[2, 1]")) {
		t.Fatal("EqualSet failed")
	}
}

// ---------------------------------------------------------------------------

