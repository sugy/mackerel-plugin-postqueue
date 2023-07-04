package mppostqueue

import "testing"

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
