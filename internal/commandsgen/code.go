package commandsgen

import (
	"bytes"
	"embed"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/temporalio/cli/internal/commandsgen/types"
)

//go:embed types/*.go
var typesFS embed.FS

func GenerateCommandsCode(pkg string, contextType string, commands Commands) ([]byte, error) {
	w := &codeWriter{allCommands: commands.CommandList, OptionSets: commands.OptionSets, contextType: contextType}

	// Generate optionsContext interface
	w.writeLinef("// optionsContext provides context for building option flags.")
	w.writeLinef("type optionsContext interface {")
	w.writeLinef("\tBindFlagEnvVar(flag *%v.Flag, envVar string)", w.importPflag())
	w.writeLinef("}\n")

	// Put terminal check at top
	w.writeLinef("var hasHighlighting = %v.IsTerminal(%v.Stdout.Fd())", w.importIsatty(), w.importPkg("os"))

	// Write all option sets (skip external ones)
	for _, optSet := range commands.OptionSets {
		if optSet.IsExternal() {
			continue
		}
		if err := optSet.writeCode(w); err != nil {
			return nil, fmt.Errorf("failed writing command %v: %w", optSet.Name, err)
		}
	}

	// Write all commands
	for _, cmd := range commands.CommandList {
		if err := cmd.writeCode(w); err != nil {
			return nil, fmt.Errorf("failed writing command %v: %w", cmd.FullName, err)
		}
	}

	// Append embedded Go files from types/ (parse imports with go/ast, write code after imports)
	err := fs.WalkDir(typesFS, "types", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || strings.Contains(path, "_test.go") {
			return err
		}
		src, err := typesFS.ReadFile(path)
		if err != nil {
			return err
		}
		return w.appendGoSource(string(src))
	})
	if err != nil {
		return nil, err
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
	allCommands []Command
	OptionSets  []OptionSets
	contextType string
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

// appendGoSource parses a Go source file, registers its imports, and appends
// everything after the import block to the output buffer.
func (c *codeWriter) appendGoSource(src string) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, parser.ImportsOnly)
	if err != nil {
		return fmt.Errorf("failed to parse embedded source: %w", err)
	}

	// Register imports
	for _, imp := range f.Imports {
		// imp.Path.Value includes quotes, so trim them
		c.importPkg(strings.Trim(imp.Path.Value, `"`))
	}

	// Find end of imports and append the rest
	var lastImportEnd token.Pos
	for _, decl := range f.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
			if genDecl.End() > lastImportEnd {
				lastImportEnd = genDecl.End()
			}
		}
	}

	// Write everything after imports
	c.buf.WriteString(src[fset.Position(lastImportEnd).Offset:])
	return nil
}

func (c *codeWriter) importPflag() string { return c.importPkg("github.com/spf13/pflag") }

func (c *codeWriter) importIsatty() string { return c.importPkg("github.com/mattn/go-isatty") }

func (c *Command) structName() string { return namify(c.FullName, true) + "Command" }

func (o *OptionSets) writeCode(w *codeWriter) error {
	if o.Name == "" {
		return fmt.Errorf("missing option set name")
	}

	// write struct
	w.writeLinef("type %v struct {", o.setStructName())
	for _, opt := range o.Options {
		if err := opt.writeStructField(w); err != nil {
			return fmt.Errorf("failed writing option set %v: %w", opt.Name, err)
		}

	}
	// Add FlagSet for tracking which flags were explicitly set
	w.writeLinef("FlagSet *%v.FlagSet", w.importPflag())
	w.writeLinef("}\n")

	// write flags
	w.writeLinef("func (v *%v) BuildFlags(ctx optionsContext, f *%v.FlagSet) {",
		o.setStructName(), w.importPflag())
	w.writeLinef("v.FlagSet = f")
	o.writeFlagBuilding("v", "f", w)
	w.writeLinef("}\n")

	return nil
}

func (c *Command) writeCode(w *codeWriter) error {
	// Find parent command if it exists
	var parent Command
	var hasParent bool
	for _, maybeParent := range w.allCommands {
		if c.isSubCommand(&maybeParent) {
			parent = maybeParent
			hasParent = true
			break
		}
	}

	// Every command is an exposed struct with the cobra command field and each
	// flag as a field on the struct
	w.writeLinef("type %v struct {", c.structName())
	if hasParent {
		w.writeLinef("Parent *%v", parent.structName())
	}
	w.writeLinef("Command %v.Command", w.importCobra())

	// Include option sets
	for _, optSetName := range c.OptionSets {
		// Find the option set to check if it's external
		var optSet *OptionSets
		for i := range w.OptionSets {
			if w.OptionSets[i].Name == optSetName {
				optSet = &w.OptionSets[i]
				break
			}
		}
		if optSet != nil && optSet.IsExternal() {
			// External option-set: import package and use its type
			pkgRef := w.importPkg(optSet.ExternalPackage)
			w.writeLinef("%v.%vOptions", pkgRef, namify(optSetName, true))
		} else {
			w.writeLinef("%vOptions", namify(optSetName, true))
		}
	}

	// Each option
	for _, opt := range c.Options {
		if err := opt.writeStructField(w); err != nil {
			return fmt.Errorf("failed writing options: %w", err)
		}
	}
	w.writeLinef("}\n")

	// Constructor builds the struct and sets the flags
	if hasParent {
		w.writeLinef("func New%v(cctx %s, parent *%v) *%v {",
			c.structName(), w.contextType, parent.structName(), c.structName())
	} else {
		w.writeLinef("func New%v(cctx %s) *%v {", c.structName(), w.contextType, c.structName())
	}
	w.writeLinef("var s %v", c.structName())
	if hasParent {
		w.writeLinef("s.Parent = parent")
	}
	// Collect subcommands
	var subCommands []Command
	for _, maybeSubCmd := range w.allCommands {
		if maybeSubCmd.isSubCommand(c) {
			subCommands = append(subCommands, maybeSubCmd)
		}
	}
	// Set basic command values
	if len(subCommands) == 0 {
		w.writeLinef("s.Command.DisableFlagsInUseLine = true")
		w.writeLinef("s.Command.Use = %q", c.NamePath[len(c.NamePath)-1]+" [flags]")
	} else {
		w.writeLinef("s.Command.Use = %q", c.NamePath[len(c.NamePath)-1])
	}
	w.writeLinef("s.Command.Short = %q", c.Summary)
	if c.DescriptionHighlighted != c.DescriptionPlain {
		w.writeLinef("if hasHighlighting {")
		w.writeLinef("s.Command.Long = %q", c.DescriptionHighlighted)
		w.writeLinef("} else {")
		w.writeLinef("s.Command.Long = %q", c.DescriptionPlain)
		w.writeLinef("}")
	} else {
		w.writeLinef("s.Command.Long = %q", c.DescriptionPlain)
	}
	if c.MaximumArgs > 0 {
		w.writeLinef("s.Command.Args = %v.MaximumNArgs(%v)", w.importCobra(), c.MaximumArgs)
	} else if c.ExactArgs > 0 {
		w.writeLinef("s.Command.Args = %v.ExactArgs(%v)", w.importCobra(), c.ExactArgs)
	} else {
		w.writeLinef("s.Command.Args = %v.NoArgs", w.importCobra())
	}
	if c.IgnoreMissingEnv {
		w.writeLinef("s.Command.Annotations = make(map[string]string)")
		w.writeLinef("s.Command.Annotations[\"ignoresMissingEnv\"] = \"true\"")
	}
	if c.Deprecated != "" {
		w.writeLinef("s.Command.Deprecated = %q", c.Deprecated)
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
	var flagAliases [][]string

	for _, opt := range c.Options {
		// Add aliases
		for _, alias := range opt.Aliases {
			flagAliases = append(flagAliases, []string{alias, opt.Name})
		}

		if err := opt.writeFlagBuilding("s", flagVar, w); err != nil {
			return fmt.Errorf("failed building option flags: %w", err)
		}
	}

	for _, include := range c.OptionSets {
		// Find include
		var optSet *OptionSets
		for i := range w.OptionSets {
			if w.OptionSets[i].Name == include {
				optSet = &w.OptionSets[i]
				for _, opt := range optSet.Options {
					for _, alias := range opt.Aliases {
						flagAliases = append(flagAliases, []string{alias, opt.Name})
					}
				}
				break
			}
		}

		if optSet != nil && optSet.IsExternal() {
			// External option-set: use type name with Options suffix
			w.writeLinef("s.%vOptions.BuildFlags(cctx, %v)", namify(include, true), flagVar)
		} else {
			// Internal option-set: use struct name
			w.writeLinef("s.%v.BuildFlags(cctx, %v)", setStructName(include), flagVar)
		}
	}

	// Generate normalize for aliases
	if len(flagAliases) > 0 {
		sort.Slice(flagAliases, func(i, j int) bool { return flagAliases[i][0] < flagAliases[j][0] })
		w.writeLinef("%v.SetNormalizeFunc(aliasNormalizer(map[string]string{", flagVar)
		for _, aliases := range flagAliases {
			w.writeLinef("%q: %q,", aliases[0], aliases[1])
		}
		w.writeLinef("}))")
	}
	// If there are no subcommands, or if subcommands are optional, we need a run function
	if len(subCommands) == 0 || c.SubcommandsOptional {
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

func (o *OptionSets) setStructName() string { return namify(o.Name, true) + "Options" }

func setStructName(name string) string { return namify(name, true) + "Options" }

func (o *OptionSets) writeFlagBuilding(selfVar, flagVar string, w *codeWriter) error {
	for _, option := range o.Options {
		if err := option.writeFlagBuilding(selfVar, flagVar, w); err != nil {
			return fmt.Errorf("failed writing flag building for option %v: %w", option.Name, err)
		}
	}
	return nil
}

func (o *Option) fieldName() string { return namify(o.Name, true) }

func (o *Option) writeStructField(w *codeWriter) error {
	var goDataType string
	switch o.Type {
	case "bool", "int", "string":
		goDataType = o.Type
	case "float":
		goDataType = "float32"
	case "duration":
		goDataType = "Duration"
	case "timestamp":
		goDataType = "Timestamp"
	case "string[]":
		goDataType = "[]string"
	case "string-enum":
		goDataType = "StringEnum"
	case "string-enum[]":
		goDataType = "StringEnumArray"
	default:
		return fmt.Errorf("unrecognized data type %v", o.Type)
	}
	w.writeLinef("%v %v", o.fieldName(), goDataType)
	return nil
}

func (o *Option) writeFlagBuilding(selfVar, flagVar string, w *codeWriter) error {
	var flagMeth, defaultLit, setDefault string
	switch o.Type {
	case "bool":
		flagMeth, defaultLit = "BoolVar", ", false"
		if o.Default != "" {
			return fmt.Errorf("cannot have default for bool var")
		}
	case "duration":
		flagMeth, setDefault = "Var", "0"
		if o.Default != "" {
			dur, err := types.ParseDuration(o.Default)
			if err != nil {
				return fmt.Errorf("invalid default: %w", err)
			}
			// We round to the nearest ms
			setDefault = fmt.Sprintf("Duration(%v * %v.Millisecond)", dur.Milliseconds(), w.importPkg("time"))
		}
	case "timestamp":
		if o.Default != "" {
			return fmt.Errorf("default value not allowed for timestamp")
		}
		flagMeth, defaultLit = "Var", ""
	case "int":
		flagMeth, defaultLit = "IntVar", ", "+o.Default
		if o.Default == "" {
			defaultLit = ", 0"
		}
	case "float":
		flagMeth, defaultLit = "Float32Var", ", "+o.Default
		if o.Default == "" {
			defaultLit = ", 0"
		}
	case "string":
		flagMeth, defaultLit = "StringVar", fmt.Sprintf(", %q", o.Default)
	case "string[]":
		if o.Default != "" {
			return fmt.Errorf("default value not allowed for string array")
		}
		flagMeth, defaultLit = "StringArrayVar", ", nil"
	case "string-enum":
		if len(o.EnumValues) == 0 {
			return fmt.Errorf("missing enum values")
		}
		// Create enum
		pieces := make([]string, len(o.EnumValues)+len(o.HiddenLegacyValues))
		for i, enumVal := range o.EnumValues {
			pieces[i] = fmt.Sprintf("%q", enumVal)
		}
		for i, legacyVal := range o.HiddenLegacyValues {
			pieces[i+len(o.EnumValues)] = fmt.Sprintf("%q", legacyVal)
		}

		w.writeLinef("%v.%v = NewStringEnum([]string{%v}, %q)",
			selfVar, o.fieldName(), strings.Join(pieces, ", "), o.Default)
		flagMeth = "Var"
	case "string-enum[]":
		if len(o.EnumValues) == 0 {
			return fmt.Errorf("missing enum values")
		}
		// Create enum
		pieces := make([]string, len(o.EnumValues)+len(o.HiddenLegacyValues))
		for i, enumVal := range o.EnumValues {
			pieces[i] = fmt.Sprintf("%q", enumVal)
		}
		for i, legacyVal := range o.HiddenLegacyValues {
			pieces[i+len(o.EnumValues)] = fmt.Sprintf("%q", legacyVal)
		}

		if o.Default != "" {
			w.writeLinef("%v.%v = NewStringEnumArray([]string{%v}, %q)",
				selfVar, o.fieldName(), strings.Join(pieces, ", "), o.Default)
		} else {
			w.writeLinef("%v.%v = NewStringEnumArray([]string{%v}, []string{})",
				selfVar, o.fieldName(), strings.Join(pieces, ", "))
		}
		flagMeth = "Var"
	default:
		return fmt.Errorf("unrecognized data type %v", o.Type)
	}

	// If there are enums, append to desc
	desc := o.Description
	if len(o.EnumValues) > 0 {
		desc += fmt.Sprintf(" Accepted values: %s.", strings.Join(o.EnumValues, ", "))
	}
	// If required, append to desc
	if o.Required {
		desc += " Required."
	}
	// If there are aliases, append to desc
	for _, alias := range o.Aliases {
		desc += fmt.Sprintf(` Aliased as "--%v".`, alias)
	}
	// If experimental, make obvious
	if o.Experimental {
		desc += " EXPERIMENTAL."
	}

	if setDefault != "" {
		// set default before calling Var so that it stores thedefault value into the flag
		w.writeLinef("%v.%v = %v", selfVar, o.fieldName(), setDefault)
	}
	if o.Short != "" {
		w.writeLinef("%v.%vP(&%v.%v, %q, %q%v, %q)", flagVar, flagMeth, selfVar, o.fieldName(), o.Name, o.Short, defaultLit, desc)
	} else {
		w.writeLinef("%v.%v(&%v.%v, %q%v, %q)", flagVar, flagMeth, selfVar, o.fieldName(), o.Name, defaultLit, desc)
	}
	if o.DisplayType != "" {
		w.writeLinef("overrideFlagDisplayType(%v.Lookup(%q), %q)", flagVar, o.Name, o.DisplayType)
	}
	if o.Required {
		w.writeLinef("_ = %v.MarkFlagRequired(%v, %q)", w.importCobra(), flagVar, o.Name)
	}
	// Use env or implied-env for environment variable binding
	envVar := o.Env
	if envVar == "" {
		envVar = o.ImpliedEnv
	}
	if envVar != "" {
		w.writeLinef("ctx.BindFlagEnvVar(%v.Lookup(%q), %q)", flagVar, o.Name, envVar)
	}
	if o.Deprecated != "" {
		w.writeLinef("_ = %v.MarkDeprecated(%q, %q)", flagVar, o.Name, o.Deprecated)
	}
	return nil
}
