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
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

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

	os.Stdout = old
	return <-outC
}
