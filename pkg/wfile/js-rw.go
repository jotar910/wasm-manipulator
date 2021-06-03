package wfile

import (
	"os"
	"os/exec"
)

// PrintJsCode prints the javascript data to a minified javascript file.
func PrintJsCode(code, outputFilename string, flags ...CmdFlags) error {
	jsFilename := ReplaceExt(outputFilename, ".out.js")
	defer func() { logError(DeleteFile(jsFilename)) }()
	if err := WriteFile(jsFilename, code); err != nil {
		return err
	}
	if _, err := MinifyJS(jsFilename, outputFilename, flags...); err != nil {
		return err
	}
	return nil
}

// MinifyJS executes the minify executable on a javascript file.
func MinifyJS(inFilename, outFilename string, flags ...CmdFlags) (string, error) {
	args := []string{inFilename}
	cmd := exec.Command(executablePath("minify"), args...)
	file, err := os.Create(outFilename)
	if err != nil {
		return "", err
	}
	defer file.Close()
	if !hasFlag(RunCommandSilent, flags) {
		cmd.Stdout = file
		cmd.Stderr = os.Stderr
	}
	err = cmd.Run()
	if err != nil {
		return "", err
	}
	return outFilename, nil
}
