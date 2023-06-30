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
	for category, _ := range p.MsgCategories {
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
				break
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

// Do the plugin
func Do() {
	optPrefix := flag.String("metric-key-prefix", "", "Metric key prefix")
	optDebug := flag.Bool("debug", false, "Debug log level")
	optPath := flag.String("path", "/usr/sbin/postqueue", "Path to postqueue command")
	optVersion := flag.Bool("version", false, "Show version")
	flag.Parse()

	if *optVersion {
		showVersion()
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

	plugin := mp.NewMackerelPlugin(&PostqueuePlugin{
		Prefix:        *optPrefix,
		PostQueuePath: *optPath,
		PostQueueArgs: []string{"-p"},
		MsgCategories: map[string]*regexp.Regexp{
			"Connection timeout":     regexp.MustCompile(`Connection timed out`),
			"Connection refused":     regexp.MustCompile(`Connection refused`),
			"Helo command rejected":  regexp.MustCompile(`Helo command rejected: Host not found`),
			"Host not found":         regexp.MustCompile(`type=MX: Host not found, try again`),
			"Mailbox full":           regexp.MustCompile(`Mailbox full`),
			"Network is unreachable": regexp.MustCompile(`Network is unreachable`),
			"No route to host":       regexp.MustCompile(`No route to host`),
			"Over quota":             regexp.MustCompile(`The email account that you tried to reach is over quota`),
			"Relay access denied":    regexp.MustCompile(`Relay access denied`),
			// Add more log categories with corresponding regular expressions
		},
	})
	plugin.Run()
}
