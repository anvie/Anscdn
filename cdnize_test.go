package "cdnize"

import (
	"testing"
)

func TestRandString(t *testing.T) {
	rv := RandStrings(5)
	testStringNeq( t, "test RandString Not equal `abc`...", rv, "abc")
	if len(rv) != 5 {
		t.Errorf("Len should be 5 got `%d`", len(rv))
	}
	rv2 := RandStrings(6)
	testStringNeq( t, "test RandString not equal `" + rv2 + "`", rv, rv2)
}

func testStringEq(t *testing.T, msg, actual, expected string) {
	if actual != expected {
		t.Errorf("%s: `%s` != `%s`", msg, actual, expected);
	}
}

func testStringNeq(t *testing.T, msg, actual, not_expected string) {
	if actual == not_expected {
		t.Errorf("%s: `%s` == `%s`", msg, actual, not_expected)
	}
}
