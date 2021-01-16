// Copyright (C) 2016-2018 Betalo AB
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"fmt"
	"net/url"
	"testing"
)

func TestParseResourcesSuccess(t *testing.T) {
	ress := []string{
		"http://user:pass@localhost:1234/path?query=val#fragment",
		"https://user:pass@localhost:1234/path?query=val#fragment",

		"ws://user:pass@localhost:42/path?query=val#fragment",
		"wss://user:pass@localhost:42/path?query=val#fragment",

		"tcp://localhost:42",
		"tcp4://localhost:42",
		"tcp6://[::1]:42",

		"file://relative/path/to/file",
		"file://relative/path/to/file#absent",
		"file:///absolute/path/to/file",
		"file://absolute/path/to/file#absent",

		"postgres://user:pass@localhost:5432/dbname?query=val#fragment",

		"mysql://user:pass@localhost:3306/dbname?query=val#fragment",

		"kafka://localhost",
		"kafka://localhost:9092",
		"kafka://localhost:9092#topics",
		"kafka://localhost:9092#topics=t1,t2",

		"kafkas://localhost",
		"kafkas://localhost:9093",
		"kafkas://localhost:9093#topics",
		"kafkas://localhost:9093#topics=t1,t2",
		"kafkas://localhost:9093#tls=skip-verify&topics=t1,t2",

		"kafkas://user:password@localhost:9093#tls=skip-verify&topics=t1,t2",
		"kafkas://user:password@localhost:9093#tls=skip-verify&topics=t1,t2&sasl=plain",
		"kafkas://user:password@localhost:9093#tls=skip-verify&topics=t1,t2&sasl=scram-sha-256",

		"command",
		"command with args",
		"relative/path/to/command",
		"relative/path/to/command with args",
		"/absolute/path/to/command",
		"/absolute/path/to/command with args",

		"",
	}

	actualRess, err := parseResources(ress)
	if err != nil {
		t.Errorf("failed to parse resources: %#v", err.Error())
	}
	if len(actualRess) != len(ress) {
		t.Errorf("missing parsed resources")
	}
}

func TestParseResourcesFailure(t *testing.T) {
	ress := []string{
		"//foo test",

		"kafkas://localhost:9093#&tls=skipverify", // notice it should be `skip-verify`
		"kafkas://user:password@localhost:9093#&sasl=wrong",
	}

	for _, urlString := range ress {
		_, err := parseResource(urlString)
		if err == nil {
			t.Errorf("expected error parsing invalid resource '%v', but got none", urlString)
		}
	}
}

func TestParseFragment(t *testing.T) {
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
		actual := parseFragment(given)
		if len(actual) != len(expected) {
			t.Errorf("unexpected parsed fragment count %d != %d: %#v",
				len(expected), len(actual), actual)
		}
		for actualK, actualV := range actual {
			if expectedV, ok := expected[actualK]; !ok || actualV[0] != expectedV {
				t.Errorf("unexpected parsed fragment k/v pair")
			}
		}
	}
}

func TestGetOptOrDefault(t *testing.T) {
	defaultVal := "default"
	tests := map[string]string{
		"":              defaultVal,
		"=":             defaultVal, // invalid format, should be skipped
		"=ignore":       defaultVal, // invalid format, should be skipped
		"baz=1":         defaultVal, // not the entry we are looking for
		"foo":           defaultVal,
		"foo=":          defaultVal,
		"foo=bar":       "bar",
		"foo=bar&baz":   "bar",
		"foo=bar&baz=1": "bar",
		"foo=bar&foo=1": "bar", // take only the first value if multiple are set
	}
	for given, expected := range tests {
		u, _ := url.Parse(fmt.Sprintf("http://somewhere/#%v", given))
		actual := getOptOrDefault(*u, "foo", defaultVal)
		if actual != expected {
			t.Errorf("unexpected opt value: expected %v, got %v", expected, actual)
		}
	}
}
