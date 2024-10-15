package commands

import (
	"bytes"
	"io"
	"os"
	"runtime"
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

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}
