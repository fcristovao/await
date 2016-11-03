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
