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
	return <-outC
}
