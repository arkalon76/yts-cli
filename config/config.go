package config

import (
	"fmt"
	"io/fs"
	"os"

	"gopkg.in/yaml.v3"
)

const (
	USER_RWX_ONLY = 0700
)

var (
	DefaultConfiguration = &Configuration{
		Filename: "config.yaml",
		Path:     "~/.config/yts",
		Transmission: Transmission{
			User:            "<rpc_username>",
			Pass:            "<rpc_password>",
			Host:            "<rpc_hostname>",
			DestinationPath: "<path for downloads>",
		},
	}
)

type Configuration struct {
	Filename     string       `yaml:"-"`
	Path         string       `yaml:"-"`
	Transmission Transmission `yaml:"transmission"`
}
type Transmission struct {
	User            string `yaml:"user"`
	Pass            string `yaml:"pass"`
	Host            string `yaml:"host"`
	DestinationPath string `yaml:"destinationPath"`
}

func NewDefault(filename, path string) (*Configuration, error) {
	return &Configuration{
		Filename: filename,
		Path:     path,
		Transmission: Transmission{
			User:            "<rpc_username>",
			Pass:            "<rpc_password>",
			Host:            "<rpc_hostname>",
			DestinationPath: "<path for downloads>",
		},
	}, nil
}

func (c *Configuration) SaveToDisk() error {
	verifyPath(c.Filename, c.Path, true)
	filebytes, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	if err := os.WriteFile(fmt.Sprintf("%s/%s", c.Path, c.Filename), filebytes, fs.FileMode(USER_RWX_ONLY)); err != nil {
		return err
	}
	return nil
}

func (c *Configuration) ConfigExist() bool {
	if err := verifyPath(c.Filename, c.Path, false); os.IsNotExist(err) {
		return false
	}
	return true
}

func verifyPath(filename, path string, create bool) error {
	if create {
		if err := os.MkdirAll(path, fs.FileMode(USER_RWX_ONLY)); err != nil {
			return err
		}
	}
	if _, err := os.Stat(fmt.Sprintf("%s/%s", path, filename)); os.IsNotExist(err) {
		return err
	}
	return nil
}
