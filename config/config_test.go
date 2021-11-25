package config_test

import (
	"io/fs"
	"os"
	"reflect"
	"testing"

	"github.com/arkalon76/yts-cli/config"
)

func TestNewDefaultConfig(t *testing.T) {
	type args struct {
		filename string
		path     string
	}
	tests := []struct {
		name    string
		args    args
		want    *config.Configuration
		wantErr bool
	}{
		{
			name: "New default configuration",
			args: args{
				filename: "config.yaml",
				path:     "~/.config/yts",
			},
			want: config.DefaultConfiguration,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := config.NewDefault(tt.args.filename, tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDefaultConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDefaultConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfiguration_ConfigExist(t *testing.T) {
	type fields struct {
		Filename     string
		SetupFunc    func()
		Path         string
		Transmission config.Transmission
	}
	tests := []struct {
		name      string
		setupFunc func()
		fields    fields
		want      bool
	}{
		{
			name: "File does not exist",
			setupFunc: func() {
				os.Remove("./config.yaml")
			},
			fields: fields{
				Filename:     "config.yaml",
				Path:         ".",
				Transmission: config.DefaultConfiguration.Transmission,
			},
			want: false,
		},
		{
			name: "File does exist",
			setupFunc: func() {
				os.WriteFile("./config.yaml", []byte("Hello"), fs.FileMode(0755))
			},
			fields: fields{
				Filename:     "config.yaml",
				Path:         ".",
				Transmission: config.DefaultConfiguration.Transmission,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &config.Configuration{
				Filename:     tt.fields.Filename,
				Path:         tt.fields.Path,
				Transmission: tt.fields.Transmission,
			}
			tt.setupFunc()
			if got := c.ConfigExist(); got != tt.want {
				t.Errorf("Configuration.ConfigExist() = %v, want %v", got, tt.want)
			}
		})
	}
	// CLEANUP TEST FILE
	os.Remove("./config.yaml")
}

func TestConfiguration_SaveToDisk(t *testing.T) {
	type fields struct {
		Filename     string
		Path         string
		Transmission config.Transmission
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Generate default config successfully",
			fields: fields{
				Filename:     "config.yaml",
				Path:         ".",
				Transmission: config.DefaultConfiguration.Transmission,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &config.Configuration{
				Filename:     tt.fields.Filename,
				Path:         tt.fields.Path,
				Transmission: tt.fields.Transmission,
			}
			if err := c.SaveToDisk(); (err != nil) != tt.wantErr {
				t.Errorf("Configuration.SaveToDisk() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	// CLEANUP TEST FILE
	// os.Remove("./config.yaml")
}
