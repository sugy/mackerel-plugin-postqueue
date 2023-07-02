package mppostqueue

import (
	"fmt"
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

// PostqueuePluginConfig is the configuration file format
type PostqueuePluginConfig struct {
	Prefix        string
	PostQueuePath string
	MsgCategories map[string]string
}

// LoadPluginConfig loads the plugin configuration file
func (c *PostqueuePluginConfig) LoadPluginConfig(configFile string) error {
	contents, err := ioutil.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("an error occurred while loading the file: %w", err)
	}

	// Perse TOML in contents
	if _, err := toml.Decode(string(contents), &c); err != nil {
		return fmt.Errorf("an error occurred while decoding TOML format file: %w", err)
	}

	return nil
}
