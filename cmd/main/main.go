package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"

	"joao/wasm-manipulator/internal/waspect"
	"joao/wasm-manipulator/internal/wconfigs"
	"joao/wasm-manipulator/internal/wgenerator"
	"joao/wasm-manipulator/internal/wyaml"
	"joao/wasm-manipulator/pkg/wfile"
)

// main is the entry function for the execution of this command line tool,
func main() {
	var code string
	var err error

	configs := wconfigs.Get()
	ext := strings.Trim(filepath.Ext(configs.InputModule), ".")

	logrus.Infof("Reading module (%s file)", ext)
	switch ext {
	case "wasm":
		code, err = wfile.ReadWasmFile(filePath(configs.InputModule))
	case "wat":
		code, err = wfile.ReadWatFile(filePath(configs.InputModule))
	default:
		err = fmt.Errorf("unknown input file extension %q", ext)
	}
	if err != nil {
		logrus.Fatalln(err)
	}

	if ext != "wasm" && configs.OutputOriginal != "" {
		logrus.Infoln("Printing untouched wasm file")
		err = wfile.PrintWasmCode(code, filePath(configs.OutputOriginal))
		if err != nil {
			logrus.Errorf("could not print the original web assembly file: %v", err)
		}
	}

	logrus.Infoln("Reading transformations (yaml file)")
	transformation, err := wyaml.Read(filePath(configs.InputTransformation))
	if err != nil {
		logrus.Fatalln(err)
	}

	// Debug purpose. TODO: delete
	logrus.Debugf(fmt.Sprintf("Input Length: %d\n", len(code)))

	// Execute transformations on module
	output, ok := waspect.Run(code, transformation)
	if !ok {
		logrus.Infoln("Finishing execution")
		return
	}

	outputWg := new(sync.WaitGroup)

	outputWg.Add(1)
	go func() {
		// Debug purpose. TODO: delete
		logrus.Debugf(fmt.Sprintf("Output Length: %d\n", len(output.String())))
		outputWg.Done()
	}()

	outputWg.Add(1)
	go func() {
		jsCode, err := output.GenerateJsData()
		if err != nil && err != wgenerator.ErrorUnnecessary {
			logrus.Infoln("Printing javascript transformations")
			logrus.Fatalln(err)
		}
		if err == nil || output.NeedJS || wconfigs.Get().PrintJS {
			logrus.Infoln("Printing javascript transformations")
			err = wfile.PrintJsCode(jsCode, filePath(configs.OutputJavascript))
			if err != nil {
				logrus.Fatalln(err)
			}
		}
		outputWg.Done()
	}()

	logrus.Infoln("Printing web assembly transformations")
	err = wfile.PrintWasmCode(output.String(), filePath(configs.OutputModule))
	if err != nil {
		logrus.Fatalln(err)
	}
	outputWg.Wait()

	logrus.Infoln("Finishing execution")
}

// init initiates the tool configurations.
func init() {
	wconfigs.SetupViperConfigs()
	setupEnvironmentVariables()
	setupLogConfigs()
}

func setupEnvironmentVariables() {
	base := wconfigs.Get().DependenciesDir
	fromBase := func(pathSteps ...string) string {
		path := strings.Join(pathSteps, string(os.PathSeparator))
		return strings.Join([]string{strings.TrimRight(base, string(os.PathSeparator)), path}, string(os.PathSeparator))
	}
	dependencies := []string{
		fromBase("wabt"),
		fromBase("minifyjs", "bin"),
		fromBase("comby"),
	}
	dependenciesList := strings.Join(dependencies, string(os.PathListSeparator))
	err := os.Setenv("PATH", strings.Join([]string{os.Getenv("PATH"), dependenciesList}, string(os.PathListSeparator)))
	if err != nil {
		logrus.Fatal(err)
	}
}

// setupLogConfigs sets up the log configurations.
func setupLogConfigs() {
	configs := wconfigs.Get()
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05.000000000"
	customFormatter.FullTimestamp = true
	logrus.SetFormatter(customFormatter)
	if configs.Verbose {
		logrus.SetLevel(logrus.TraceLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
	if configs.LogFile != "" {
		filename := filePath(configs.LogFile)
		file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0666)
		if err == nil {
			fmt.Println("Running tool...")
			fmt.Printf("Logging to %q\n", filename)
			logrus.SetOutput(file)
		} else {
			logrus.Errorf("Failed to log to file, using default stdout/stderr")
		}
	}
}

// filePath returns the file path for some filename.
func filePath(name string) string {
	dataPath := wconfigs.Get().DataDir
	if dataPath == "" {
		return name
	}
	return fmt.Sprintf("%s/%s", strings.TrimRight(dataPath, "/"), name)
}
