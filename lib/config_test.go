package mppostqueue

import (
	"reflect"
	"testing"
)

func TestPostqueuePluginConfig_loadPluginConfig(t *testing.T) {
	type fields struct {
		Prefix        string
		PostQueuePath string
		MsgCategories map[string]string
	}
	type args struct {
		configFile string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "check config",
			fields: fields{},
			args: args{
				configFile: "../testdata/config.toml",
			},
			wantErr: false,
		},
		{
			name:   "check error loading the file",
			fields: fields{},
			args: args{
				configFile: "../testdata/dummy.toml",
			},
			wantErr: true,
		},
		{
			name:   "check error decoding toml",
			fields: fields{},
			args: args{
				configFile: "../testdata/config_fail.toml",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &PostqueuePluginConfig{}
			if err := c.loadPluginConfig(tt.args.configFile); (err != nil) != tt.wantErr {
				t.Errorf("PostqueuePluginConfig.loadPluginConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPostqueuePluginConfig_generateConfig(t *testing.T) {
	tests := []struct {
		name       string
		wantOutput []string
	}{
		{
			name: "check generating config toml",
			wantOutput: []string{
				`# Postqueue plugin config file`,
				`# Prefix for metrics`,
				`Prefix = "postfix"`,
				``,
				`# Path to postqueue command`,
				`PostQueuePath = "/usr/sbin/postqueue"`,
				``,
				`# Message categories`,
				`# Format: <category> = "<regex>"`,
				`[MsgCategories]`,
				`  "Connection refused" = "Connection refused"`,
				`  "Connection timeout" = "Connection timed out"`,
				`  "Helo command rejected" = "Helo command rejected: Host not found"`,
				`  "Host not found" = "type=MX: Host not found, try again"`,
				`  "Mailbox full" = "Mailbox full"`,
				`  "Network is unreachable" = "Network is unreachable"`,
				`  "No route to host" = "No route to host"`,
				`  "Over quota" = "The email account that you tried to reach is over quota"`,
				`  "Relay access denied" = "Relay access denied"`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &PostqueuePluginConfig{}
			if output := c.generateConfig(); !reflect.DeepEqual(output, tt.wantOutput) {
				t.Errorf("PostqueuePluginConfig.generateConfig() output = %v, wantoutput = %v, DeepEqual = %v", output, tt.wantOutput, reflect.DeepEqual(output, tt.wantOutput))
			}
		})
	}
}
