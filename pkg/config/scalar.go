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
	"errors"

	"gopkg.in/yaml.v3"
)

// GetScalarValue returns the value of a scalar property
// ex. if config yaml contains line "namespace: default", it will
// return ("default", nil) for key="namespace"
func (cfg *Config) GetScalarValue(key string) (string, error) {
	record, err := cfg.getScalarNode(key)
	if err != nil {
		return "", err
	}
	return record.Value, nil
}

// GetScalarValue sets the value of a scalar property
// ex. executing the command with properties (key="namespace", value="default")
// will create/update a line in config "namespace: default"
// return ("default", nil) for key="namespace"
func (cfg *Config) SetScalarValue(key string, value string) error {
	record, err := cfg.getScalarNode(key)
	if err != nil {
		key := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: key,
		}
		value := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: value,
		}

		cfg.Root.Content[0].Content = append(cfg.Root.Content[0].Content, key, value)
	} else if record != nil {
		record.Value = value
	}

	return writeConfig(cfg)
}

func (cfg *Config) getScalarNode(key string) (*yaml.Node, error) {
	if len(cfg.Root.Content) > 0 {
		nodes := cfg.Root.Content[0].Content
		for i, n := range nodes {
			if n.Value == key {
				var value *yaml.Node
				if i < len(nodes)-1 {
					value = nodes[i+1]
				}
				return value, nil
			}
		}
	}

	return nil, errors.New("unable to find key " + key)
}
