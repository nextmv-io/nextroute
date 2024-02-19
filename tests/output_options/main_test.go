package main

import (
	"os"
	"testing"

	"github.com/nextmv-io/sdk/golden"
)

const templateName = "nextroute"

func TestMain(m *testing.M) {
	cleanUp()
	golden.CopyFile("../../cmd/main.go", "main.go")
	code := m.Run()
	cleanUp()
	os.Exit(code)
}

// TestOptions tests showing the options repeated on the output.
func TestOptions(t *testing.T) {
	golden.BashTest(t, ".", golden.BashConfig{
		DisplayStdout: true,
		DisplayStderr: true,
	})
}

func cleanUp() {
	golden.Reset([]string{
		"testdata",
		"main_test.go",
		"main.sh",
		"main.sh.golden",
	})
}
