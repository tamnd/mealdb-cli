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

func TestClassify(t *testing.T) {
	typ, id, err := Domain{}.Classify("53000")
	if err != nil {
		t.Fatalf("Classify error: %v", err)
	}
	if typ != "meal" {
		t.Errorf("Classify type = %q, want meal", typ)
	}
	if id != "53000" {
		t.Errorf("Classify id = %q, want 53000", id)
	}
}

func TestLocate(t *testing.T) {
	got, err := Domain{}.Locate("meal", "53000")
	want := "https://www.themealdb.com/meal/53000"
	if err != nil || got != want {
		t.Errorf("Locate = (%q, %v), want (%q, nil)", got, err, want)
	}
}

func TestClassifyEmpty(t *testing.T) {
	_, _, err := Domain{}.Classify("")
	if err == nil {
		t.Error("Classify(\"\") should return error")
	}
}
