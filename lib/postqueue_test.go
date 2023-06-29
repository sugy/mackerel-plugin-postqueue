package mppostqueue

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
	"testing"

	mp "github.com/mackerelio/go-mackerel-plugin"
)

// Helper function to read a file and return its contents as a string
func readFile(filePath string) (string, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("An error occurred while opening the file: %w", err)
	}
	defer file.Close()

	// Load the file's contents to a string
	contents, err := ioutil.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("An error occurred while loading the file: %w", err)
	}

	return string(contents), nil
}

func TestPostqueuePlugin_GraphDefinition(t *testing.T) {
	type fields struct {
		MsgCategories map[string]*regexp.Regexp
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]mp.Graphs
	}{
		{
			name: "check graph definition",
			fields: fields{
				MsgCategories: map[string]*regexp.Regexp{
					"Foo": regexp.MustCompile(`^Foo\s+\-`),
				},
			},
			want: map[string]mp.Graphs{
				"postqueue": {
					Label: "Postfix postqueue",
					Unit:  "integer",
					Metrics: []mp.Metrics{
						{Name: "Foo", Label: "Foo"},
						{Name: "queue", Label: "queue"},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PostqueuePlugin{
				MsgCategories: tt.fields.MsgCategories,
			}
			if got := p.GraphDefinition(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PostqueuePlugin.GraphDefinition() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPostqueuePlugin_FetchMetrics(t *testing.T) {
	type fields struct {
		TestdataPath  string
		MsgCategories map[string]*regexp.Regexp
	}
	tests := []struct {
		name    string
		fields  fields
		want    map[string]float64
		wantErr bool
	}{
		{
			name: "check metrics",
			fields: fields{
				TestdataPath: "../testdata/postqueue_output.txt",
				MsgCategories: map[string]*regexp.Regexp{
					"Connection timeout":    regexp.MustCompile(`Connection timed out`),
					"Connection refused":    regexp.MustCompile(`Connection refused`),
					"Helo command rejected": regexp.MustCompile(`Helo command rejected: Host not found`),
					"Over quota":            regexp.MustCompile(`The email account that you tried to reach is over quota`),
				},
			},
			want: map[string]float64{
				"Connection_timeout":    4,
				"Connection_refused":    1,
				"Helo_command_rejected": 1,
				"Over_quota":            1,
				"queue":                 12,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// テスト用の出力を読み込む
			output, err := readFile(tt.fields.TestdataPath)
			if err != nil {
				t.Errorf("PostqueuePlugin.FetchMetrics() error = %v", err)
				return
			}

			p := &PostqueuePlugin{
				PostQueueOutput: output,
				MsgCategories:   tt.fields.MsgCategories,
			}
			got, err := p.FetchMetrics()
			if (err != nil) != tt.wantErr {

				t.Errorf("PostqueuePlugin.FetchMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PostqueuePlugin.FetchMetrics() = %v, want %v", got, tt.want)
			}
		})
	}
}