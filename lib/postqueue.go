package mppostqueue

import (
	"flag"
	"math/rand"
	"strings"
	"time"

	mp "github.com/mackerelio/go-mackerel-plugin"
)

// Use go-mackerel-plugin to create this postqueue plugin.
// If you want to know how to use go-mackerel-plugin library,
// see https://github.com/mackerelio/go-mackerel-plugin .

// PostqueuePlugin mackerel plugin
type PostqueuePlugin struct {
	Prefix string
}

// MetricKeyPrefix interface for PluginWithPrefix
func (p *PostqueuePlugin) MetricKeyPrefix() string {
	if p.Prefix == "" {
		p.Prefix = "postqueue"
	}
	return p.Prefix
}

// GraphDefinition interface for mackerelplugin
func (p *PostqueuePlugin) GraphDefinition() map[string]mp.Graphs {
	labelPrefix := strings.Title(p.Prefix)
	return map[string]mp.Graphs{
		"dice": {
			Label: labelPrefix + " Dice Value",
			Unit:  "integer",
			Metrics: []mp.Metrics{
				{Name: "d6", Label: "Dice(d6)"},
				{Name: "d20", Label: "Dice(d20)"},
			},
		},
	}
}

// FetchMetrics interface for mackerelplugin
func (p *PostqueuePlugin) FetchMetrics() (map[string]float64, error) {
	rand.Seed(time.Now().UnixNano())
	metrics := map[string]float64{
		"d6":  float64(rand.Intn(6) + 1),
		"d20": float64(rand.Intn(20) + 1),
	}
	return metrics, nil
}

// Do the plugin
func Do() {
	optPrefix := flag.String("metric-key-prefix", "", "Metric key prefix")
	flag.Parse()

	plugin := mp.NewMackerelPlugin(&PostqueuePlugin{
		Prefix: *optPrefix,
	})
	plugin.Run()
}
