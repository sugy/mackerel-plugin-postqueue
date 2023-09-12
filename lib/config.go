package mppostqueue

import (
	"fmt"
	"os"
	"sort"

	"github.com/BurntSushi/toml"
	log "github.com/sirupsen/logrus"
)

// PostqueuePluginConfig is the configuration file format
type PostqueuePluginConfig struct {
	Prefix        string
	PostQueuePath string
	MsgCategories map[string]string
}

// loadPluginConfig loads the plugin configuration file
func (c *PostqueuePluginConfig) loadPluginConfig(configFile string) error {
	contents, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("an error occurred while loading the file: %w", err)
	}

	// Perse TOML in contents
	if _, err := toml.Decode(string(contents), &c); err != nil {
		return fmt.Errorf("an error occurred while decoding TOML format file: %w", err)
	}

	return nil
}

// Generate config file template
func (c *PostqueuePluginConfig) generateConfig() []string {
	c.Prefix = "postfix"
	c.PostQueuePath = "/usr/sbin/postqueue"

	c.MsgCategories = getDefaultMsgCategories()
	keys := c.getMsgCategoriesKeys()
	sort.Strings(keys)
	log.Debug("generateConfig: MsgCategories keys: ", keys)

	var result []string
	result = append(result, `# Postqueue plugin config file`)

	// Output config file template
	result = append(result, `# Prefix for metrics`)
	result = append(result, `Prefix = "`+c.Prefix+`"`)
	result = append(result, ``)

	result = append(result, `# Path to postqueue command`)
	result = append(result, `PostQueuePath = "`+c.PostQueuePath+`"`)
	result = append(result, ``)

	result = append(result, `# Message categories`)
	result = append(result, `# Format: <category> = "<regex>"`)
	result = append(result, `[MsgCategories]`)
	for k := range keys {
		result = append(result, `  "`+keys[k]+`" = "`+c.MsgCategories[keys[k]]+`"`)
	}

	return result
}

// Get MsgCategories keys
func (c *PostqueuePluginConfig) getMsgCategoriesKeys() []string {
	keys := make([]string, 0, len(c.MsgCategories))
	for k := range c.MsgCategories {
		keys = append(keys, k)
	}
	return keys
}

// Set default MsgCategories
func getDefaultMsgCategories() map[string]string {
	return map[string]string{
		"Connection refused":     "Connection refused",
		"Connection timeout":     "Connection timed out",
		"Helo command rejected":  "Helo command rejected: Host not found",
		"Host not found":         "type=MX: Host not found, try again",
		"Mailbox full":           "Mailbox full",
		"Network is unreachable": "Network is unreachable",
		"No route to host":       "No route to host",
		"Over quota":             "The email account that you tried to reach is over quota",
		"Relay access denied":    "Relay access denied",
		// Add more log categories with corresponding regular expressions
	}
}
