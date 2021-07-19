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
	"strings"

	"gopkg.in/yaml.v3"
)

// SetSequenceValue sets an entry for sequence property
// ex. of sequence in .yml:
// mySequence: # root
//   - key: myKey1 # entry 1
//     value: myValue1
//   - key: myKey2 # entry 2
//     value: myValue2
//
// key property must follow format <suquenceName>.<entryKeyValue>, ex "mySequence.myKey1"
func (cfg *Config) SetSequenceValue(key string, value string) error {
	seqKey, entryKey, err := splitKey(key)
	if err != nil {
		return err
	}

	if err = setSequenceRoot(cfg, seqKey); err != nil {
		return err
	}

	seqRoot, err := cfg.getScalarNode(seqKey)
	if err != nil {
		return err
	}

	if seqRoot.Kind == yaml.ScalarNode {
		// if node is empty, it will be read as scalar, not sequence.
		// set node back to sequence
		seqRoot.Kind = yaml.SequenceNode
		seqRoot.Tag = "!!seq"
	}

	entry, err := cfg.getSequenceEntryNode(seqKey, entryKey)
	if err != nil {
		entry = createSequenceEntry(entryKey, value)
		seqRoot.Content = append(seqRoot.Content, entry)
	} else {
		if len(entry.Content) < 4 {
			return errors.New("unable to update entry value for key " + key)
		}
		entry.Content[3].Value = value // index 3 is the custom value node, ex: "value: myValue"
	}

	return writeConfig(cfg)
}

func setSequenceRoot(cfg *Config, seqKey string) error {
	_, err := cfg.getScalarNode(seqKey)
	if err != nil {
		seqRoot := &yaml.Node{
			Value: seqKey,
			Kind:  yaml.ScalarNode,
		}
		seq := &yaml.Node{
			Kind: yaml.SequenceNode,
		}
		cfg.Root.Content[0].Content = append(cfg.Root.Content[0].Content, seqRoot, seq)
	}
	return nil
}

func (cfg *Config) getSequenceEntryNode(seqKey string, entryKey string) (*yaml.Node, error) {
	root, err := cfg.getScalarNode(seqKey)
	if err != nil {
		return nil, err
	}

	seqEntries := root.Content
	for _, e := range seqEntries {
		if len(e.Content) < 2 {
			continue
		}

		eKey := e.Content[1].Value // index 1 is the key name, ex: "key: myKeyName"

		if strings.Compare(entryKey, eKey) == 0 {
			return e, nil
		}
	}

	return nil, errors.New("unable to find key " + entryKey)
}

func createSequenceEntry(key string, value string) *yaml.Node {
	keyN := []*yaml.Node{{
		Kind:  yaml.ScalarNode,
		Value: "key",
	}, {
		Kind:  yaml.ScalarNode,
		Value: key,
	}}

	valueN := []*yaml.Node{{
		Kind:  yaml.ScalarNode,
		Value: "value",
	}, {
		Kind:  yaml.ScalarNode,
		Value: value,
	}}

	entry := &yaml.Node{
		Kind:    yaml.MappingNode,
		Content: append(keyN, valueN...),
	}

	return entry
}

func isSequenceKey(key string) bool {
	parts := strings.Split(key, ".")
	return len(parts) == 2
}

func splitKey(key string) (string, string, error) {
	if !isSequenceKey(key) {
		return "", "", errors.New("sequence key format should follow <key>.<subkey>")
	}

	parts := strings.Split(key, ".")
	return parts[0], parts[1], nil
}
