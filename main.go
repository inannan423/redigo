package main

import (
	// Add import
	"fmt"
	"os"
	"path/filepath" // Add import
	"redigo/config"
	"redigo/lib/logger"
	"redigo/resp/handler"
	"redigo/tcp"
)

// Default configuration file name
const defaultConfigFileName string = "redis.conf" // Modify constant name

var defaultProperties = &config.ServerProperties{
	Bind: "0.0.0.0",
	Port: 6379,
}

// Command line argument for specifying config file path
var configPath string // Add variable

// func init() {
// 	// Add command line argument support, allowing users to specify config file via -c flag
// 	flag.StringVar(&configPath, "c", "", "Config file path (e.g., /path/to/redis.conf)")
// 	flag.Parse()
// }

// fileExists remains as is or use as needed
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	return err == nil && !info.IsDir()
}

// Added: function to find configuration file
func findConfigFile() string {
	// 1. First priority: use path specified by command line argument
	if configPath != "" {
		if fileExists(configPath) {
			return configPath
		} else {
			fmt.Printf("Warning: Config file specified by command line does not exist: %s\n", configPath)
		}
	}

	// 2. Try to find in binary executable directory
	execPath, err := os.Executable() // Get current executable path
	if err == nil {
		execDir := filepath.Dir(execPath)                              // Get directory of executable
		pathInExecDir := filepath.Join(execDir, defaultConfigFileName) // Join paths
		if fileExists(pathInExecDir) {
			return pathInExecDir // Return path if found
		}
	}

	// 3. Try to find in current working directory (as last resort)
	if fileExists(defaultConfigFileName) {
		return defaultConfigFileName
	}

	// If not found in any location, return empty string
	return ""
}

func main() {
	logger.Setup(&logger.Settings{
		Path:       "logs",
		Name:       "redigo",
		Ext:        "log",
		TimeFormat: "2006-01-02",
	})

	// Modified: call new function to determine config file path
	configFileToLoad := findConfigFile()

	if configFileToLoad != "" { // If config file is found
		fmt.Printf("Loading config file: %s\n", configFileToLoad)
		config.SetupConfig(configFileToLoad) // Load config using found path
	} else {
		fmt.Printf("Config file not found in standard locations, using default config\n")
		config.Properties = defaultProperties // Use default configuration
	}

	err := tcp.ListenAndServeWithSignal(
		&tcp.Config{
			Address: fmt.Sprintf("%s:%d",
				config.Properties.Bind,
				config.Properties.Port),
		},
		handler.MakeHandler())
	if err != nil {
		logger.Error(err)
	}
}
