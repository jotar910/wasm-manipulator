package wfile

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// ReadWatFile reads wat file returning the module content.
func ReadWatFile(watFilename string, flags ...CmdFlags) (string, error) {
	wasmFilename, err := ConvertWatToWasm(watFilename, flags...)
	defer func() { logError(DeleteFile(wasmFilename)) }()
	if err != nil {
		return "", fmt.Errorf("converting temporary file %q to binary: %w", watFilename, err)
	}
	return ReadWasmFile(wasmFilename, flags...)
}

// ReadWasmFile reads wasm file returning the module content.
func ReadWasmFile(wasmFilename string, flags ...CmdFlags) (string, error) {
	watFilename, err := ConvertWasmToWat(wasmFilename, flags...)
	defer func() { logError(DeleteFile(watFilename)) }()
	if err != nil {
		return "", fmt.Errorf("converting temporary file %q to textual: %w", wasmFilename, err)
	}
	return ReadFile(watFilename)
}

// PrintWasmCode prints a module to a wat file.
func PrintWasmCode(code, outputFilename string) error {
	watFilename, err := SaveWat(code)
	defer func() { logError(DeleteFile(watFilename)) }()
	if err != nil {
		return fmt.Errorf("saving web assembly textual code: %w", err)
	}
	if _, err := ConvertWatToWasmFile(watFilename, outputFilename); err != nil {
		return err
	}
	return nil
}

// logError logs an error to the output logger.
func logError(err error) {
	if err != nil {
		logrus.Error(err)
	}
}
