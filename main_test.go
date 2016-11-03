package main

import (
	"flag"
	"os"
	"testing"
)

func TestSplitArgsResourceAndCommand(t *testing.T) {
	os.Args = []string{"./await", "true", "--", "echo"}
	flag.Parse()
	res, cmd := splitArgs(flag.Args())
	if len(res) != 1 || res[0] != "true" {
		t.Errorf("no resources found")
	}
	if len(cmd) != 1 || cmd[0] != "echo" {
		t.Errorf("no command found")
	}
}

func TestSplitArgsResourceAndNoCommand(t *testing.T) {
	os.Args = []string{"./await", "true", "--"}
	flag.Parse()
	res, cmd := splitArgs(flag.Args())
	if len(res) != 1 || res[0] != "true" {
		t.Errorf("no resources found")
	}
	if len(cmd) != 0 {
		t.Errorf("unexpected command found")
	}

	os.Args = []string{"./await", "true"}
	flag.Parse()
	res, cmd = splitArgs(flag.Args())
	if len(res) != 1 || res[0] != "true" {
		t.Errorf("no resources found")
	}
	if len(cmd) != 0 {
		t.Errorf("unexpected command found")
	}
}

func TestSplitArgsNoResourceAndCommand(t *testing.T) {
	os.Args = []string{"./await", "--", "echo"}
	flag.Parse()
	res, cmd := splitArgs(flag.Args())
	if len(res) != 0 {
		t.Errorf("unexpected resource found")
	}
	if len(cmd) != 1 || cmd[0] != "echo" {
		t.Errorf("no command found")
	}
}

func TestSplitArgsNoResourceNoCommand(t *testing.T) {
	os.Args = []string{"./await", "--"}
	flag.Parse()
	res, cmd := splitArgs(flag.Args())
	if len(res) != 0 {
		t.Errorf("unexpected resource found")
	}
	if len(cmd) != 0 {
		t.Errorf("unexpected command found")
	}

	os.Args = []string{"./await"}
	flag.Parse()
	res, cmd = splitArgs(flag.Args())
	if len(res) != 0 {
		t.Errorf("unexpected resource found")
	}
	if len(cmd) != 0 {
		t.Errorf("unexpected command found")
	}
}

func TestParseTags(t *testing.T) {
	tests := map[string]map[string]string{
		"":              map[string]string{},
		"=":             map[string]string{}, // invalid format, should be skipped
		"=ignore":       map[string]string{}, // invalid format, should be skipped
		"foo":           map[string]string{"foo": ""},
		"foo=":          map[string]string{"foo": ""},
		"foo=bar":       map[string]string{"foo": "bar"},
		"foo=bar&baz":   map[string]string{"foo": "bar", "baz": ""},
		"foo=bar&baz=1": map[string]string{"foo": "bar", "baz": "1"},
	}
	for given, expected := range tests {
		actual := parseTags(given)
		if len(actual) != len(expected) {
			t.Errorf("unexpected parsed tags count %d != %d: %#v",
				len(expected), len(actual), actual)
		}
		for actualK, actualV := range actual {
			if expectedV, ok := expected[actualK]; !ok || actualV != expectedV {
				t.Errorf("unexpected parsed tags k/v pair")
			}
		}
	}
}
