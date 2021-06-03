package wfile

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

type CmdFlags int

const (
	RunCommandSilent CmdFlags = iota >> 1
	RunCommandNoFoldWatExpr
)

// SaveWat saves wat code in a wat file.
func SaveWat(code string) (string, error) {
	filename, err := TempFileName("", ".wat")
	if err != nil {
		return "", err
	}
	err = WriteFile(filename, code)
	if err != nil {
		return "", err
	}
	return filename, nil
}

// ConvertWatToWasm converts wat file into wasm file.
func ConvertWatToWasm(inFilename string, flags ...CmdFlags) (string, error) {
	outFilename := ReplaceExt(inFilename, ".wasm")
	return ConvertWatToWasmFile(inFilename, outFilename, flags...)
}

// ConvertWatToWasm converts wat file into wasm file.
func ConvertWatToWasmFile(inFilename, outFilename string, flags ...CmdFlags) (string, error) {
	args := []string{inFilename, "--no-check", "-o", outFilename}
	cmd := exec.Command(executablePath("wat2wasm"), args...)
	if !hasFlag(RunCommandSilent, flags) {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return outFilename, nil
}

// ConvertWatToWasm converts wasm file into wat file.
func ConvertWasmToWat(inFilename string, flags ...CmdFlags) (string, error) {
	outFilename := ReplaceExt(inFilename, ".wat")
	argsFlags := []string{inFilename, "--no-check", "--generate-names"}
	if !hasFlag(RunCommandNoFoldWatExpr, flags) {
		argsFlags = append(argsFlags, "--fold-exprs")
	}
	args := append(argsFlags, []string{"-o", outFilename}...)
	cmd := exec.Command(executablePath("wasm2wat"), args...)
	if !hasFlag(RunCommandSilent, flags) {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return outFilename, nil
}

// WriteFile writes content to file.
func WriteFile(filename, content string) error {
	err := ioutil.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("writing to file %q: %w", filename, err)
	}
	return nil
}

// ReadFile reads string content from file.
func ReadFile(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("opening file %q: %w", filename, err)
	}
	defer f.Close()
	out, err := ioutil.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("reading from file %q: %w", filename, err)
	}
	return string(out), nil
}

// DeleteFile deletes file by name.
func DeleteFile(filename string) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return err
	}
	err := os.Remove(filename)
	if err != nil {
		logrus.Errorf("not removed: remove file error: %v", err)
		return err
	}
	return nil
}

// TempFileName generates a temporary filename.
func TempFileName(prefix, suffix string) (string, error) {
	randBytes := make([]byte, 16)
	if _, err := rand.Read(randBytes); err != nil {
		return "", fmt.Errorf("generate random temporary filename: %w", err)
	}
	return filepath.Join(os.TempDir(), prefix+hex.EncodeToString(randBytes)+suffix), nil
}

// TempDir creates a temporary directory.
func TempDir(content string) (string, error) {
	d, err := ioutil.TempDir("", "")
	if err != nil {
		return "", err
	}
	f, err := ioutil.TempFile(d, "")
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err = f.Write([]byte(content)); err != nil {
		return "", err
	}
	return d, nil
}

// ReplaceExt replaces filename extension for another.
func ReplaceExt(inFilename, prefix string) string {
	ext := path.Ext(inFilename)
	return inFilename[0:len(inFilename)-len(ext)] + prefix
}

// executablePath returns the name with the executable path.
func executablePath(name string) string {
	return name
}

// hasFlag returns if finds a flag.
func hasFlag(flag CmdFlags, flags []CmdFlags) bool {
	for _, f := range flags {
		if f&flag != 0 {
			return true
		}
	}
	return false
}
