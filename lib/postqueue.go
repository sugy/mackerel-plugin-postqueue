package mppostqueue

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	mp "github.com/mackerelio/go-mackerel-plugin"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// PostqueuePlugin mackerel plugin for Postfix postqueue metrics
type PostqueuePlugin struct {
	Prefix          string
	PostQueuePath   string
	PostQueueArgs   []string
	PostQueueOutput string
	MsgCategories   map[string]*regexp.Regexp
}

// MetricKeyPrefix interface for PluginWithPrefix
func (p *PostqueuePlugin) MetricKeyPrefix() string {
	if p.Prefix == "" {
		p.Prefix = "postfix"
	}
	return p.Prefix
}

// GraphDefinition interface for mackerelplugin
func (p *PostqueuePlugin) GraphDefinition() map[string]mp.Graphs {
	labelPrefix := cases.Title(language.Und, cases.NoLower).String(p.Prefix)
	metrics := []mp.Metrics{}

	// Add metrics for Message categories
	for category := range p.MsgCategories {
		name := strings.Replace(category, " ", "_", -1)
		metrics = append(metrics, mp.Metrics{Name: name, Label: category})
	}
	metrics = append(metrics, mp.Metrics{Name: "queue", Label: "queue"})

	label := "Postfix postqueue"
	if len(labelPrefix) > 0 {
		label = labelPrefix + " postqueue"
	}

	return map[string]mp.Graphs{
		"postqueue": {
			Label:   label,
			Unit:    "integer",
			Metrics: metrics,
		},
	}
}

// FetchMetrics interface for mackerelplugin
func (p *PostqueuePlugin) FetchMetrics() (map[string]float64, error) {
	if p.PostQueueOutput == "" {
		output, err := p.runPostQueueCommand()
		if err != nil {
			return nil, err
		}
		log.Debug("FetchMetrics (output): ", fmt.Sprintf("'%v'", output))
		p.PostQueueOutput = output
	}

	// Initialize metric map
	metrics := make(map[string]float64)
	// Initialize metric map for Message categories
	for category := range p.MsgCategories {
		name := strings.Replace(category, " ", "_", -1)
		metrics[name] = 0
	}

	// Read and classify output entries
	scanner := bufio.NewScanner(strings.NewReader(p.PostQueueOutput))
	for scanner.Scan() {
		line := scanner.Text()
		for category, regex := range p.MsgCategories {
			if regex.MatchString(line) {
				log.Debug("FetchMetrics (line): ", fmt.Sprintf("'%v'", line))
				name := strings.Replace(category, " ", "_", -1)
				metrics[name] = metrics[name] + 1
				// not break here, because one line may match multiple categories
			}
		}
		// line の先頭が 10桁以上の16進数であれば、それはキューIDとみなす
		if len(line) >= 10 && regexp.MustCompile(`^[0-9A-F]{10,12}[*]{0,1}\s+`).MatchString(line) {
			metrics["queue"] = metrics["queue"] + 1
		}
	}
	log.Debug("FetchMetrics (metrics): ", fmt.Sprintf("'%v'", metrics))
	return metrics, nil
}

// runPostQueueCommand executes the "postqueue -p" command and returns its output
func (p *PostqueuePlugin) runPostQueueCommand() (string, error) {
	log.Debug(fmt.Sprintf("command: %v %v", p.PostQueuePath, strings.Join(p.PostQueueArgs, " ")))
	cmd := exec.Command(p.PostQueuePath, p.PostQueueArgs...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := cmd.ProcessState.ExitCode()

	if err != nil {
		return "", fmt.Errorf(fmt.Sprintf("failed to execute postqueue command. exitCode: %d, Stdout: '%s', Stderr: '%s'\n", exitCode, stdout.String(), stderr.String()))
	}
	return stdout.String(), err
}

// loadPluginConfig loads config file
func (p *PostqueuePlugin) loadPluginConfig(configFile string) error {
	c := &PostqueuePluginConfig{}
	// Load config file
	err := c.loadPluginConfig(configFile)
	if err != nil {
		return err
	}

	// Set config file values
	if c.Prefix != "" {
		p.Prefix = c.Prefix
	}
	if c.PostQueuePath != "" {
		p.PostQueuePath = c.PostQueuePath
	}

	// Set config file values for Message categories
	if c.MsgCategories != nil {
		p.MsgCategories = getMsgCategories(c.MsgCategories)
	}

	return nil
}

// getMsgCategories returns the map of message categories
func getMsgCategories(m map[string]string) map[string]*regexp.Regexp {
	msgCategories := make(map[string]*regexp.Regexp)
	for category, regex := range m {
		if category != "" && regex != "" {
			msgCategories[category] = regexp.MustCompile(regex)
		}
	}
	return msgCategories
}

func (p *PostqueuePlugin) validate() error {
	if p.PostQueuePath == "" {
		return fmt.Errorf("postqueue path is required")
	}
	if len(p.MsgCategories) == 0 {
		return fmt.Errorf("message categories is required")
	}
	return nil
}

// Generate config file template
func generateConfig() {
	c := &PostqueuePluginConfig{}
	c.generateConfig()
}

// Do the plugin
func Do() {
	optPrefix := flag.String("metric-key-prefix", "", "Metric key prefix")
	optDebug := flag.Bool("debug", false, "Debug log level")
	optPath := flag.String("path", "/usr/sbin/postqueue", "Path to postqueue command")
	optVersion := flag.Bool("version", false, "Show version")
	optConfig := flag.String("config", "", "Path to TOML format config file")
	optGenerateConfig := flag.Bool("generate-config", false, "Generate config file template")
	flag.Parse()

	if *optVersion {
		showVersion()
		os.Exit(0)
	}

	if *optGenerateConfig {
		generateConfig()
		os.Exit(0)
	}

	customFmt := new(log.TextFormatter)
	customFmt.TimestampFormat = "2006-01-02 15:04:05"
	customFmt.FullTimestamp = true
	log.SetFormatter(customFmt)
	log.SetOutput(os.Stdout)

	if *optDebug {
		log.SetLevel(log.DebugLevel)
	}

	p := &PostqueuePlugin{}
	p.PostQueueArgs = []string{"-p"}

	if *optConfig != "" {
		err := p.loadPluginConfig(*optConfig)
		if err != nil {
			log.Errorf("Failed to load config file: %s", err)
			os.Exit(1)
		}
	}

	// Set command line values
	if *optPrefix != "" {
		p.Prefix = *optPrefix
	}
	if *optPath != "" {
		p.PostQueuePath = *optPath
	}

	// Set default values for Message categories
	if p.MsgCategories == nil {
		p.MsgCategories = getMsgCategories(getDefaultMsgCategories())
	}

	log.Debug("Do (p): ", fmt.Sprintf("'%v'", p))

	err := p.validate()
	if err != nil {
		log.Errorf("Failed to validate config: %s", err)
		os.Exit(1)
	}

	plugin := mp.NewMackerelPlugin(p)
	plugin.Run()
}
