// The MIT License
//
// Copyright (c) 2020 Temporal Technologies Inc.  All rights reserved.
//
// Copyright (c) 2020 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func readConfig() (*Config, error) {
	path, err := configFile()
	if err != nil {
		return nil, err
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var cfgNode yaml.Node
	err = yaml.Unmarshal(data, &cfgNode)
	if err != nil {
		return nil, err
	}

	cfg := &Config{Root: &cfgNode}

	err = setRootIfEmpty(cfg)

	return cfg, err
}
func writeConfig(cfg *Config) error {
	data, err := yaml.Marshal(cfg.Root)
	if err != nil {
		return err
	}

	fpath, err := configFile()
	if err != nil {
		return err
	}

	file, err := os.OpenFile(fpath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(0666))
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}

func configFile() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	fpath := filepath.Join(dir, "tctl.yml")

	if _, err := os.Stat(fpath); err != nil {
		fmt.Printf("creating config file: %v\n", fpath)
		file, err := os.Create(fpath)
		if err != nil {
			defer file.Close()
			return fpath, err
		}
	}

	return fpath, nil
}

func configDir() (string, error) {
	dpath, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	dpath = filepath.Join(dpath, ".config", "temporalio")

	if _, err := os.Stat(dpath); err != nil {
		fmt.Printf("creating dir: %v\n", dpath)
		err = os.MkdirAll(dpath, 0755)
		return dpath, err
	}

	return dpath, nil
}

func setRootIfEmpty(cfg *Config) error {
	if cfg.Root.Kind == yaml.Kind(0) || len(cfg.Root.Content) == 0 {
		cfg.Root.Kind = yaml.DocumentNode
		cfg.Root.Content = []*yaml.Node{{Kind: yaml.MappingNode}}
	}

	return nil
}
