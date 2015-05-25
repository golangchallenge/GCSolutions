package config

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

// Config provides access to the application configuration
type Config struct {
	vars map[string]string
	sync.Mutex
}

// New create a new Config pointer and loads the configuration from a given JSON file
func New(filename string) *Config {
	config := Config{}
	config.Setup(filename)
	return &config
}

// Get looks up a config value for a given key and returns it.
func (c *Config) Get(key string) string {
	c.Lock()
	defer c.Unlock()
	if _, ok := c.vars[key]; !ok {
		panic(fmt.Sprintf("Missing config key: %s.", key))
	}
	return c.vars[key]
}

// Set sets config value for a given key.
func (c *Config) Set(key, value string) {
	c.Lock()
	defer c.Unlock()
	c.vars[key] = value
}

// Setup loads the configuration from a given JSON file.
func (c *Config) Setup(path string) {
	c.Lock()
	defer c.Unlock()
	createConfigFile(path)
	c.vars = readConfigVars(path)
}

// readConfigVars opens a given file, decodes the content from JSON and returns the resulting map.
func readConfigVars(configFile string) map[string]string {
	file, _ := os.Open(configFile)
	decoder := json.NewDecoder(file)
	vars := make(map[string]string)
	err := decoder.Decode(&vars)
	if err != nil {
		panic(err.Error())
	}
	for key := range vars {
		if os.Getenv(key) != "" {
			vars[key] = os.Getenv(key)
		}
	}
	return vars
}

// createConfigFile checks if a given config file exist, and if not creates it
// from a file with the same name and suffixed by ".dist".
func createConfigFile(configFile string) {
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		dist := configFile + ".dist"
		if _, err := os.Stat(dist); os.IsNotExist(err) {
			panic(fmt.Sprintf("Missing config file at %s.", dist))
		}
		err := copyFile(dist, configFile)
		if err != nil {
			panic(fmt.Sprintf("Unable to create config at %s.", configFile))
		}
		log.Printf("Created config file at %s.", configFile)
	}
}

// copyFile copies a given source file to a given destination.
func copyFile(src, dst string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()
	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(d, s); err != nil {
		d.Close()
		return err
	}
	return d.Close()
}
