package db

import "testing"

func TestFixNumber(t *testing.T) {
	s1 := "7145"
	res1 := fixNumber(s1)
	if res1 != 71.45 {
		t.Fatalf("%f is not 71.45\n", res1)
	}

	s2 := "714.5"
	res2 := fixNumber(s2)
	if res2 != 71.45 {
		t.Fatalf("%f is not 71.45\n", res2)
	}
}
