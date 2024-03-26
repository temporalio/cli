package commandsmd

import (
	"bytes"
	"fmt"
	"go/format"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"
)

func GenerateCommandsCode(pkg string, commands []*Command) ([]byte, error) {
	w := &codeWriter{allCommands: commands}
	// Put terminal check at top
	w.writeLinef("var hasHighlighting = %v.IsTerminal(%v.Stdout.Fd())", w.importIsatty(), w.importPkg("os"))

	// Write all commands, then come back and write package and imports
	for _, cmd := range commands {
		if err := cmd.writeCode(w); err != nil {
			return nil, fmt.Errorf("failed writing command %v: %w", cmd.FullName, err)
		}
	}

	// Write package and imports to final buf
	var finalBuf bytes.Buffer
	finalBuf.WriteString("// Code generated. DO NOT EDIT.\n\n")
	finalBuf.WriteString("package " + pkg + "\n\nimport(\n")
	// Sort imports before writing
	importLines := make([]string, 0, len(w.imports))
	for _, v := range w.imports {
		importLines = append(importLines, fmt.Sprintf("%q\n", v))
	}
	sort.Strings(importLines)
	for _, v := range importLines {
		finalBuf.WriteString(v + "\n")
	}
	finalBuf.WriteString(")\n\n")
	_, _ = finalBuf.ReadFrom(&w.buf)

	// Format and return
	b, err := format.Source(finalBuf.Bytes())
	if err != nil {
		err = fmt.Errorf("failed generating code: %w, code:\n-----\n%s\n-----", err, finalBuf.Bytes())
	}
	return b, err
}

type codeWriter struct {
	buf         bytes.Buffer
	allCommands []*Command
	// Key is short ref, value is full
	imports map[string]string
}

var regexNonAlnum = regexp.MustCompile("[^A-Za-z0-9]+")

func namify(s string, capitalizeFirst bool) string {
	// Split on every non-alnum
	ret := ""
	for i, piece := range regexNonAlnum.Split(s, -1) {
		if i > 0 || capitalizeFirst {
			piece = strings.ToUpper(piece[:1]) + piece[1:]
		}
		ret += piece
	}
	return ret
}

func (c *codeWriter) writeLinef(s string, args ...any) {
	// Ignore errors
	_, _ = c.buf.WriteString(fmt.Sprintf(s, args...) + "\n")
}

func (c *codeWriter) importPkg(pkg string) string {
	// For now we'll just panic on dupe and assume last path element is pkg name
	ref := strings.TrimPrefix(path.Base(pkg), "go-")
	if prev := c.imports[ref]; prev == "" {
		if c.imports == nil {
			c.imports = make(map[string]string)
		}
		c.imports[ref] = pkg
	} else if prev != pkg {
		panic(fmt.Sprintf("duplicate import for %v and %v", pkg, prev))
	}
	return ref
}

func (c *codeWriter) importCobra() string { return c.importPkg("github.com/spf13/cobra") }

func (c *codeWriter) importPflag() string { return c.importPkg("github.com/spf13/pflag") }

func (c *codeWriter) importIsatty() string { return c.importPkg("github.com/mattn/go-isatty") }

func (c *Command) structName() string { return namify(c.FullName, true) + "Command" }

func (c *Command) isSubCommand(maybeParent *Command) bool {
	return len(c.NamePath) == len(maybeParent.NamePath)+1 && strings.HasPrefix(c.FullName, maybeParent.FullName+" ")
}

func (c *Command) writeCode(w *codeWriter) error {
	// Add every named options set as a separate struct
	for _, optSet := range c.OptionsSets {
		if optSet.SetName == "" {
			continue
		}
		// Struct
		w.writeLinef("type %v struct {", optSet.setStructName())
		if err := optSet.writeStructFields(w); err != nil {
			return fmt.Errorf("failed writing option set %v: %w", optSet.SetName, err)
		}
		w.writeLinef("}\n")
		// Flag builder
		w.writeLinef("func (v *%v) buildFlags(cctx *CommandContext, f *%v.FlagSet) {",
			optSet.setStructName(), w.importPflag())
		optSet.writeFlagBuilding("v", "f", w)
		w.writeLinef("}\n")
	}

	// Find parent command if it exists
	var parent *Command
	for _, maybeParent := range w.allCommands {
		if c.isSubCommand(maybeParent) {
			parent = maybeParent
			break
		}
	}

	// Every command is an exposed struct with the cobra command field and each
	// flag as a field on the struct
	w.writeLinef("type %v struct {", c.structName())
	if parent != nil {
		w.writeLinef("Parent *%v", parent.structName())
	}
	w.writeLinef("Command %v.Command", w.importCobra())
	for _, optSet := range c.OptionsSets {
		// Includes as embedded
		for _, include := range optSet.IncludeOptionsSets {
			w.writeLinef("%vOptions", namify(include, true))
		}
		// If there is a set name, it is treated as embedded because options sets
		// are different structs, so the fields are not set
		if optSet.SetName != "" {
			w.writeLinef("%v", optSet.setStructName())
			continue
		}
		if err := optSet.writeStructFields(w); err != nil {
			return fmt.Errorf("failed writing options: %w", err)
		}
	}
	w.writeLinef("}\n")

	// Constructor builds the struct and sets the flags
	if parent != nil {
		w.writeLinef("func New%v(cctx *CommandContext, parent *%v) *%v {",
			c.structName(), parent.structName(), c.structName())
	} else {
		w.writeLinef("func New%v(cctx *CommandContext) *%v {", c.structName(), c.structName())
	}
	w.writeLinef("var s %v", c.structName())
	if parent != nil {
		w.writeLinef("s.Parent = parent")
	}
	// Collect subcommands
	var subCommands []*Command
	for _, maybeSubCmd := range w.allCommands {
		if maybeSubCmd.isSubCommand(c) {
			subCommands = append(subCommands, maybeSubCmd)
		}
	}
	// Set basic command values
	if len(subCommands) == 0 {
		w.writeLinef("s.Command.DisableFlagsInUseLine = true")
		w.writeLinef("s.Command.Use = %q", c.NamePath[len(c.NamePath)-1]+" [flags]"+c.UseSuffix)
	} else {
		w.writeLinef("s.Command.Use = %q", c.NamePath[len(c.NamePath)-1])
	}
	w.writeLinef("s.Command.Short = %q", c.Short)
	if c.LongHighlighted != c.LongPlain {
		w.writeLinef("if hasHighlighting {")
		w.writeLinef("s.Command.Long = %q", c.LongHighlighted)
		w.writeLinef("} else {")
		w.writeLinef("s.Command.Long = %q", c.LongPlain)
		w.writeLinef("}")
	} else {
		w.writeLinef("s.Command.Long = %q", c.LongPlain)
	}
	if c.MaximumArgs > 0 {
		w.writeLinef("s.Command.Args = %v.MaximumNArgs(%v)", w.importCobra(), c.MaximumArgs)
	} else if c.ExactArgs > 0 {
		w.writeLinef("s.Command.Args = %v.ExactArgs(%v)", w.importCobra(), c.ExactArgs)
	} else {
		w.writeLinef("s.Command.Args = %v.NoArgs", w.importCobra())
	}
	// Add subcommands
	for _, subCommand := range subCommands {
		w.writeLinef("s.Command.AddCommand(&New%v(cctx, &s).Command)", subCommand.structName())
	}
	// Set flags
	flagVar := "s.Command.Flags()"
	if len(subCommands) > 0 {
		// If there are subcommands, this needs to be persistent flags
		flagVar = "s.Command.PersistentFlags()"
	}
	for _, optSet := range c.OptionsSets {
		// If there's a name, this is done in the method
		if optSet.SetName != "" {
			w.writeLinef("s.%v.buildFlags(cctx, %v)", optSet.setStructName(), flagVar)
			continue
		}
		// Each field
		if err := optSet.writeFlagBuilding("s", flagVar, w); err != nil {
			return fmt.Errorf("failed building option flags: %w", err)
		}
	}
	// If there are no subcommands, we need a run function
	if len(subCommands) == 0 {
		w.writeLinef("s.Command.Run = func(c *%v.Command, args []string) {", w.importCobra())
		w.writeLinef("if err := s.run(cctx, args); err != nil {")
		w.writeLinef("cctx.Options.Fail(err)")
		w.writeLinef("}")
		w.writeLinef("}")
	}
	// Init
	if c.HasInit {
		w.writeLinef("s.initCommand(cctx)")
	}
	w.writeLinef("return &s")
	w.writeLinef("}\n")
	return nil
}

func (c *CommandOptions) setStructName() string { return namify(c.SetName, true) + "Options" }

func (c *CommandOptions) writeStructFields(w *codeWriter) error {
	for _, option := range c.Options {
		if err := option.writeStructField(w); err != nil {
			return fmt.Errorf("failed writing struct field for option %v: %w", option.Name, err)
		}
	}
	return nil
}

func (c *CommandOptions) writeFlagBuilding(selfVar, flagVar string, w *codeWriter) error {
	// Embedded sets
	for _, include := range c.IncludeOptionsSets {
		w.writeLinef("%v.%vOptions.buildFlags(cctx, %v)", selfVar, namify(include, true), flagVar)
	}
	// Each direct option
	for _, option := range c.Options {
		if err := option.writeFlagBuilding(selfVar, flagVar, w); err != nil {
			return fmt.Errorf("failed writing flag building for option %v: %w", option.Name, err)
		}
	}
	return nil
}

func (c *CommandOption) fieldName() string { return namify(c.Name, true) }

func (c *CommandOption) writeStructField(w *codeWriter) error {
	var goDataType string
	switch c.DataType {
	case "bool", "int", "string":
		goDataType = c.DataType
	case "duration":
		goDataType = w.importPkg("time") + ".Duration"
	case "timestamp":
		goDataType = "Timestamp"
	case "string[]":
		goDataType = "[]string"
	case "string-enum":
		goDataType = "StringEnum"
	default:
		return fmt.Errorf("unrecognized data type %v", c.DataType)
	}
	w.writeLinef("%v %v", c.fieldName(), goDataType)
	return nil
}

func (c *CommandOption) writeFlagBuilding(selfVar, flagVar string, w *codeWriter) error {
	var flagMeth, defaultLit string
	switch c.DataType {
	case "bool":
		flagMeth, defaultLit = "BoolVar", ", false"
		if c.DefaultValue != "" {
			return fmt.Errorf("cannot have default for bool var")
		}
	case "duration":
		flagMeth, defaultLit = "DurationVar", ", 0"
		if c.DefaultValue != "" {
			dur, err := time.ParseDuration(c.DefaultValue)
			if err != nil {
				return fmt.Errorf("invalid default: %w", err)
			}
			// We round to the nearest ms
			defaultLit = fmt.Sprintf(", %v * %v.Millisecond", int64(dur/time.Millisecond), w.importPkg("time"))
		}
	case "timestamp":
		if c.DefaultValue != "" {
			return fmt.Errorf("default value not allowed for timestamp")
		}
		flagMeth, defaultLit = "Var", ""
	case "int":
		flagMeth, defaultLit = "IntVar", ", "+c.DefaultValue
		if c.DefaultValue == "" {
			defaultLit = ", 0"
		}
	case "string":
		flagMeth, defaultLit = "StringVar", fmt.Sprintf(", %q", c.DefaultValue)
	case "string[]":
		if c.DefaultValue != "" {
			return fmt.Errorf("default value not allowed for string array")
		}
		flagMeth, defaultLit = "StringArrayVar", ", nil"
	case "string-enum":
		if len(c.EnumValues) == 0 {
			return fmt.Errorf("missing enum values")
		}
		// Create enum
		pieces := make([]string, len(c.EnumValues))
		for i, enumVal := range c.EnumValues {
			pieces[i] = fmt.Sprintf("%q", enumVal)
		}
		w.writeLinef("%v.%v = NewStringEnum([]string{%v}, %q)",
			selfVar, c.fieldName(), strings.Join(pieces, ", "), c.DefaultValue)
		flagMeth = "Var"
	default:
		return fmt.Errorf("unrecognized data type %v", c.DataType)
	}

	// If there are enums, append to desc
	desc := c.Desc
	if len(c.EnumValues) > 0 {
		desc += fmt.Sprintf(" Accepted values: %s.", strings.Join(c.EnumValues, ", "))
	}

	if c.Alias == "" {
		w.writeLinef("%v.%v(&%v.%v, %q%v, %q)",
			flagVar, flagMeth, selfVar, c.fieldName(), c.Name, defaultLit, desc)
	} else {
		w.writeLinef("%v.%vP(&%v.%v, %q, %q%v, %q)",
			flagVar, flagMeth, selfVar, c.fieldName(), c.Name, c.Alias, defaultLit, desc)
	}
	if c.Required {
		w.writeLinef("_ = %v.MarkFlagRequired(%v, %q)", w.importCobra(), flagVar, c.Name)
	}
	if c.EnvVar != "" {
		w.writeLinef("cctx.BindFlagEnvVar(%v.Lookup(%q), %q)", flagVar, c.Name, c.EnvVar)
	}
	if c.Hidden {
		w.writeLinef("%v.Lookup(%q).Hidden = true", flagVar, c.Name)
	}
	return nil
}
