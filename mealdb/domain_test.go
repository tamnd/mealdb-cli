package mealdb

import (
	"testing"
)

// These tests are offline: they exercise the URI driver's pure string functions,
// which need no network. The client's HTTP behaviour is covered in mealdb_test.go.

func TestDomainInfo(t *testing.T) {
	info := Domain{}.Info()
	if info.Scheme != "mealdb" {
		t.Errorf("Scheme = %q, want mealdb", info.Scheme)
	}
	if len(info.Hosts) == 0 || info.Hosts[0] != Host {
		t.Errorf("Hosts = %v, want [%s]", info.Hosts, Host)
	}
	if info.Identity.Binary != "mealdb" {
		t.Errorf("Identity.Binary = %q, want mealdb", info.Identity.Binary)
	}
}

func TestClassifyNumericIsID(t *testing.T) {
	typ, id, err := Domain{}.Classify("52959")
	if err != nil {
		t.Fatalf("Classify error: %v", err)
	}
	if typ != "id" {
		t.Errorf("Classify type = %q, want id", typ)
	}
	if id != "52959" {
		t.Errorf("Classify id = %q, want 52959", id)
	}
}

func TestClassifyStringIsQuery(t *testing.T) {
	typ, id, err := Domain{}.Classify("pasta")
	if err != nil {
		t.Fatalf("Classify error: %v", err)
	}
	if typ != "query" {
		t.Errorf("Classify type = %q, want query", typ)
	}
	if id != "pasta" {
		t.Errorf("Classify id = %q, want pasta", id)
	}
}

func TestClassifyEmpty(t *testing.T) {
	_, _, err := Domain{}.Classify("")
	if err == nil {
		t.Error("Classify(\"\") should return error")
	}
}

func TestLocateID(t *testing.T) {
	got, err := Domain{}.Locate("id", "52959")
	want := "https://www.themealdb.com/meal/52959"
	if err != nil || got != want {
		t.Errorf("Locate = (%q, %v), want (%q, nil)", got, err, want)
	}
}

func TestLocateQuery(t *testing.T) {
	got, err := Domain{}.Locate("query", "52959")
	want := "https://www.themealdb.com/meal/52959-Detail.php"
	if err != nil || got != want {
		t.Errorf("Locate = (%q, %v), want (%q, nil)", got, err, want)
	}
}

func TestLocateUnknownType(t *testing.T) {
	_, err := Domain{}.Locate("bogus", "123")
	if err == nil {
		t.Error("Locate with unknown type should return error")
	}
}

func TestIsNumericHelpers(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"52959", true},
		{"0", true},
		{"pasta", false},
		{"53abc", false},
		{"", false},
		{"  ", false},
	}
	for _, tc := range cases {
		got := isNumeric(tc.input)
		if got != tc.want {
			t.Errorf("isNumeric(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}
