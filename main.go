// Copyright (c) 2016 Betalo AB
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
	"flag"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"
)

func main() {
	var (
		forceFlag    = flag.Bool("f", false, "Force running the command even after giving up")
		quietFlag    = flag.Bool("q", false, "Set quiet mode")
		timeoutFlag  = flag.Duration("t", 1*time.Minute, "Set timeout duration before giving up")
		verbose1Flag = flag.Bool("v", false, "Set verbose output mode")
		verbose2Flag = flag.Bool("vv", false, "Set more verbose output mode")
	)
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: await [options...] <res>... [ -- <cmd>]")
		fmt.Fprintln(os.Stderr, "Await availability of resources.")
		fmt.Fprintln(os.Stderr)
		flag.PrintDefaults()
	}
	flag.Parse()

	var logLevel int
	switch {
	case *quietFlag:
		logLevel = silentLevel
	case *verbose1Flag:
		logLevel = infoLevel
	case *verbose2Flag:
		logLevel = debugLevel
	default:
		logLevel = errorLevel
	}

	logger := NewLogger(logLevel)
	resArgs, cmdArgs := splitArgs(flag.Args())

	ress, err := parseResources(resArgs)
	if err != nil {
		logger.Fatalln("Error: failed to parse resources: %v", err)
	}

	awaiter := &awaiter{
		logger:  logger,
		timeout: *timeoutFlag,
	}

	if err := awaiter.run(ress); err != nil {
		if e, ok := err.(*unavailabilityError); ok {
			logger.Errorf("Resource unavailable: %v", e)
			logger.Errorln("Timeout exceeded")
		} else {
			logger.Fatalln("Error: %v", err)
		}
		if !*forceFlag {
			os.Exit(1)
		}
	} else {
		logger.Infoln("All resources available")
	}

	if len(cmdArgs) > 0 {
		logger.Infof("Runnning command: %v", cmdArgs)
		if err := execCmd(cmdArgs); err != nil {
			logger.Fatalf("Error: failed to execute command: %v", err)
		}
	}
}

func splitArgs(args []string) ([]string, []string) {
	if i := indexOf(args, "--"); i >= 0 {
		return args[0:i], args[i+1:]
	}

	// We haven't seen a resource|command separator ('--'). This can either be
	// because of Go's flag parser removing the separator if no args were given,
	// or because there actually was none given.
	// Fallback to the original (unparsed) flag argument list and see if a
	// separator was given there and if, assume all arguments given are part of
	// the command, i.e. no resources at all were provided.
	if i := indexOf(os.Args, "--"); i >= 0 {
		return []string{}, args
	}

	// We still haven't seen a resource|command separator ('--'). Now finally
	// assume because there actually was none given and use all arguments as
	// resources.
	return args, []string{}
}

func indexOf(l []string, s string) int {
	for i, e := range l {
		if e == "--" {
			return i
		}
	}
	return -1
}

func execCmd(cmdArgs []string) error {
	path, err := exec.LookPath(cmdArgs[0])
	if err != nil {
		return err
	}
	return syscall.Exec(path, cmdArgs, os.Environ())
}
