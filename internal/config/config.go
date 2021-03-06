package config

import (
	"fmt"
	"io/ioutil"
	"os/user"
	"path/filepath"

	"github.com/go-pkg-org/gopkg/internal/util/file"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
)

// GoPkgDir is the directory where gopkg meta files are placed.
const GoPkgDir = ".gopkg"

const configFile = ".gopkg.yaml"

// Maintainer is the object containg info about the maintainer.
type Maintainer struct {
	Email string `yaml:"email" envconfig:"email"`
	Name  string `yaml:"name" envconfig:"name"`
}

// Config is the root object containg the configuration file.
type Config struct {
	BinDir     string     `yaml:"bin_dir" envconfig:"bin_dir"`
	CachePath  string     `yaml:"cache_path" envconfig:"cache_path"`
	Maintainer Maintainer `yaml:"maintainer" envconfig:"maintainer"`
	SrcDir     string     `yaml:"src_dir"  envconfig:"src_dir"`
}

// Load loads the configuration file from the users home directory.
func (c *Config) load() error {
	u, err := user.Current()
	if err != nil {
		return err
	}

	path, err := file.FindByExtensions(filepath.Join(u.HomeDir, configFile), []string{"yaml", "yml"})
	if err == nil {
		out, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		err = yaml.Unmarshal([]byte(out), &c)
		if err != nil {
			return err
		}
	}

	err = envconfig.Process("gopkg", c)
	if err != nil {
		return err
	}

	return nil
}

// Default returns a default configuration.
func Default() (*Config, error) {
	u, err := user.Current()
	if err != nil {
		return nil, err
	}

	c := &Config{
		BinDir:    filepath.Join(u.HomeDir, GoPkgDir, "bin"),
		CachePath: filepath.Join(u.HomeDir, GoPkgDir, "cache.json"),
		SrcDir:    filepath.Join(u.HomeDir, GoPkgDir, "src"),
	}

	if err := c.load(); err != nil {
		return nil, err
	}

	return c, nil
}

// GetGoPathDir returns GOPATH variable
func (c *Config) GetGoPathDir() (string, error) {
	return filepath.Join(c.SrcDir, ".."), nil
}

// GetMaintainerEntry returns the maintainer entry: format Name <Email>
func (c *Config) GetMaintainerEntry() string {
	return fmt.Sprintf("%s <%s>", c.Maintainer.Name, c.Maintainer.Email)
}
