package commandsgen

import (
	"bytes"
	"fmt"
	"strings"
)

// TODO: figure out how to generate each file:
//
//	activity, batch, cmd-options, env, index,
//	operator, schedule, server, task-queue, workflow
// cmd-options -
//     - tags: not sure where this comes from

// index doesn't need to be generated

type DocsFile struct {
	FileName string
}

func GenerateDocsFiles(commands Commands) (map[string][]byte, error) {
	// Fix CRLF
	// md := bytes.ReplaceAll(DocsYAML, []byte("\r\n"), []byte("\n"))
	// var m Docs
	// err := yaml.Unmarshal(md, &m)
	// if err != nil {
	// 	return Commands{}, fmt.Errorf("failed unmarshalling yaml: %w", err)
	// }

	w := &docWriter{fileMap: make(map[string]*bytes.Buffer)}

	// sort by parent command (activity, batch, etc)
	for _, cmd := range commands.CommandList {
		if err := cmd.writeDoc(w); err != nil {
			return nil, fmt.Errorf("failed writing docs for command %s: %w", cmd.FullName, err)
		}
	}

	// Write package and imports to final buf
	// var finalBuf bytes.Buffer
	// finalBuf.WriteString("// Code generated. DO NOT EDIT.\n\n")

	// Format and return
	var finalMap = make(map[string][]byte)
	for key, buf := range w.fileMap {
		// b, err := format.Source(buf.Bytes())
		// if err != nil {
		// 	return nil, fmt.Errorf("failed formatting docs: %w, docs:\n-----\n%s\n-----", err, buf.Bytes())
		// }

		finalMap[key] = buf.Bytes()
	}
	return finalMap, nil
}

type docWriter struct {
	fileMap map[string]*bytes.Buffer
}

// func (c *docWriter) writeLinef(s string, args ...any) {
// 	// Ignore errors
// 	_, _ = c.buf.WriteString(fmt.Sprintf(s, args...) + "\n")
// }

func (c *Command) writeDoc(w *docWriter) error {
	// If this is a root command, write a new file
	if len(strings.Split(c.FullName, " ")) == 2 {
		w.writeCommand(c)
	} else if len(strings.Split(c.FullName, " ")) > 2 {
		w.writeSubcommand(c)
	}
	return nil
}

func (w *docWriter) writeCommand(c *Command) error {
	fileName := strings.Split(c.FullName, " ")[1]
	w.fileMap[fileName] = &bytes.Buffer{}
	w.fileMap[fileName].WriteString("---\n")
	w.fileMap[fileName].WriteString("id: " + fileName + "\n")
	w.fileMap[fileName].WriteString("title: " + c.FullName + "\n")
	w.fileMap[fileName].WriteString("sidebar_label: " + c.FullName + "\n")
	w.fileMap[fileName].WriteString("description: " + c.Description + "\n")
	w.fileMap[fileName].WriteString("toc_max_heading_level: 4\n")
	w.fileMap[fileName].WriteString("keywords:\n") // TODO
	w.fileMap[fileName].WriteString("tags:\n")     // TODO
	w.fileMap[fileName].WriteString("---\n\n")
	return nil
}

func (w *docWriter) writeSubcommand(c *Command) error {
	// TODO: write options from command, parent command, and global command
	fileName := strings.Split(c.FullName, " ")[1]
	subCommand := strings.Join(strings.Split(c.FullName, " ")[2:], "")
	w.fileMap[fileName].WriteString("## " + subCommand + "\n")
	return nil
}
