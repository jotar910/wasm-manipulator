package wconfigs

import (
	"joao/wasm-manipulator/pkg/wfile"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	ConfigInModule        = "in_module"
	ConfigInTransform     = "in_transform"
	ConfigOutModule       = "out_module"
	ConfigOutJavascript   = "out_js"
	ConfigOutModuleOrig   = "out_module_orig"
	ConfigDependenciesDir = "dependencies_dir"
	ConfigDataDir         = "data_dir"
	ConfigLogFile         = "log_file"
	ConfigInclude         = "include"
	ConfigExclude         = "exclude"
	ConfigPrintJS         = "print_js"
	ConfigAllowEmpty      = "allow_empty"
	ConfigVerbose         = "verbose"
	ConfigIgnoreOrder     = "ignore_order"
)

var (
	defaults = map[string]interface{}{
		ConfigInModule:        "input.wasm",
		ConfigInTransform:     "input.yml",
		ConfigOutModule:       "output.wasm",
		ConfigOutJavascript:   "",
		ConfigOutModuleOrig:   "",
		ConfigDependenciesDir: "",
		ConfigDataDir:         "",
		ConfigLogFile:         "",
		ConfigInclude:         nil,
		ConfigExclude:         nil,
		ConfigPrintJS:         false,
		ConfigAllowEmpty:      false,
		ConfigVerbose:         false,
		ConfigIgnoreOrder:     false,
	}
)

var config ToolConfig

// ToolConfig represents the configurations model for the tool.
type ToolConfig struct {
	InputModule         string   `mapstructure:"in_module"`
	InputTransformation string   `mapstructure:"in_transform"`
	OutputModule        string   `mapstructure:"out_module"`
	OutputJavascript    string   `mapstructure:"out_js"`
	OutputOriginal      string   `mapstructure:"out_module_orig"`
	DependenciesDir     string   `mapstructure:"dependencies_dir"`
	DataDir             string   `mapstructure:"data_dir"`
	LogFile             string   `mapstructure:"log_file"`
	Include             []string `mapstructure:"include"`
	Exclude             []string `mapstructure:"exclude"`
	PrintJS             bool     `mapstructure:"print_js"`
	AllowEmpty          bool     `mapstructure:"allow_empty"`
	Verbose             bool     `mapstructure:"verbose"`
	ConfigIgnoreOrder   bool     `mapstructure:"ignore_order"`
}

// Get returns the tool configurations.
func Get() ToolConfig {
	return config
}

// SetupViperConfigs sets up the viper configurations.
func SetupViperConfigs() {
	// Update runtime defaults.
	baseAux, err := os.Getwd()
	if err != nil {
		logrus.Fatal(err)
	}
	base := strings.TrimRight(baseAux, string(os.PathSeparator))
	defaults[ConfigDependenciesDir] = strings.Join([]string{base, "dependencies"}, string(os.PathSeparator))

	viper.SetEnvPrefix("wmr")

	for k, v := range defaults {
		viper.SetDefault(k, v)
		err := viper.BindEnv(k)
		if err != nil {
			logrus.Fatalf("could not bind environment variable to viper: %v", err)
		}
	}

	pflag.String(ConfigInModule, viper.GetString(ConfigInModule), "input filename with the module")
	pflag.String(ConfigInTransform, viper.GetString(ConfigInTransform), "YAML input filename with the transformation")
	pflag.String(ConfigOutJavascript, viper.GetString(ConfigOutJavascript), "JavaScript output filename for the results")
	pflag.String(ConfigOutModule, viper.GetString(ConfigOutModule), "WASM output filename for the results")
	pflag.String(ConfigOutModuleOrig, viper.GetString(ConfigOutModuleOrig), "WASM output filename for the original module")
	pflag.String(ConfigDependenciesDir, viper.GetString(ConfigDependenciesDir), "directory path for the dependencies data")
	pflag.String(ConfigDataDir, viper.GetString(ConfigDataDir), "directory path for the files data")
	pflag.String(ConfigLogFile, viper.GetString(ConfigLogFile), "filename to output log messages")
	pflag.StringSlice(ConfigInclude, viper.GetStringSlice(ConfigInclude), "filter which advices are included in the execution")
	pflag.StringSlice(ConfigExclude, viper.GetStringSlice(ConfigExclude), "filter which advices are included from the execution")
	pflag.Bool(ConfigPrintJS, viper.GetBool(ConfigPrintJS), "always print the auxiliar javascript code (by default only prints if necessary)")
	pflag.Bool(ConfigAllowEmpty, viper.GetBool(ConfigAllowEmpty), "allow execution even with no advices found (applies global transformations)")
	pflag.Bool(ConfigVerbose, viper.GetBool(ConfigVerbose), "the tool is executed in verbose mode")
	pflag.Bool(ConfigIgnoreOrder, viper.GetBool(ConfigIgnoreOrder), "skips the advice order field")
	pflag.Parse()

	err = viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		logrus.Fatalf("could not bind command line flags to viper: %v", err)
	}

	config = ToolConfig{}
	err = viper.Unmarshal(&config)
	if err != nil {
		logrus.Fatalf("could not decode config into struct: %v", err)
	}
	if config.OutputJavascript == "" {
		config.OutputJavascript = wfile.ReplaceExt(config.OutputModule, ".js")
	}
}
