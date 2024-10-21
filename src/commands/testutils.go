package commands

import (
	"bytes"
	"io"
	"os"
	"runtime"
	"sync"
)

var MockGOOS string

func GetGOOS() string {
	if MockGOOS != "" {
		return MockGOOS
	}
	return runtime.GOOS
}

func CaptureOutput(f func()) string {
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()

	os.Stdout = w
	os.Stderr = w

	outC := make(chan string)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		var buf bytes.Buffer
		wg.Done()
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	wg.Wait()
	f()
	w.Close()

	os.Stdout = oldStdout
	os.Stderr = oldStderr
	out := <-outC

	return out
}

// CaptureOutputAndError New function to capture both stdout and stderr separately
func CaptureOutputAndError(f func()) (string, string) {
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()

	os.Stdout = wOut
	os.Stderr = wErr

	outC := make(chan string)
	errC := make(chan string)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		var buf bytes.Buffer
		wg.Done()
		io.Copy(&buf, rOut)
		outC <- buf.String()
	}()
	go func() {
		var buf bytes.Buffer
		wg.Done()
		io.Copy(&buf, rErr)
		errC <- buf.String()
	}()

	wg.Wait()
	f()
	wOut.Close()
	wErr.Close()

	os.Stdout = oldStdout
	os.Stderr = oldStderr
	stdout := <-outC
	stderr := <-errC

	return stdout, stderr
}
