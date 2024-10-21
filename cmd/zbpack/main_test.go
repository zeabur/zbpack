package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var binName = "zbpack"

func TestMain(m *testing.M) {
	fmt.Println("Running build...")

	tempPath, err := os.MkdirTemp("", "")
	if err != nil {
		fmt.Println("Cannot get absolute path")
		os.Exit(1)
	}

	binName = tempPath + "/" + binName

	if runtime.GOOS == "windows" {
		binName += ".exe"
	}

	build := exec.Command("go", "build", "-o", binName)

	if err := build.Run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Cannot build tool %s: %s", binName, err)
		os.Exit(1)
	}

	fmt.Println("Running test...")
	result := m.Run()

	fmt.Println("Cleaning up")
	_ = os.RemoveAll(tempPath)

	os.Exit(result)
}

func TestCLI(t *testing.T) {
	t.Run("show help information", func(t *testing.T) {
		cmd := exec.Command(binName, "-h")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(string(out), "Usage") {
			t.Fatal("expected help output, but got: ", string(out))
		}
	})

	t.Run("show error when not give path", func(t *testing.T) {
		cmd := exec.Command(binName)
		out, err := cmd.CombinedOutput()
		if err == nil {
			t.Fatal("expected error, but got nil")
		}

		if !strings.Contains(string(out), "Error") {
			t.Fatal("expected error output, but got: ", string(out))
		}
	})

	t.Run("only show info when give --info flag", func(t *testing.T) {
		path, _ := os.Getwd()
		path = filepath.Join(path, "../../")
		testFilePath := filepath.Join(path, "/tests/nodejs-a-lot-of-dependencies")
		cmd := exec.Command(binName, "--info", testFilePath)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(string(out), "nodejs") {
			t.Fatal("expected info output, but got: ", string(out))
		}
	})
}
