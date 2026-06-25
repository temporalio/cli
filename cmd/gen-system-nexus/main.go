package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/temporalio/cli/internal/systemnexusgen"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

type stringSlice []string

func (s *stringSlice) String() string { return strings.Join(*s, ",") }
func (s *stringSlice) Set(v string) error {
	*s = append(*s, v)
	return nil
}

func run() error {
	var (
		pkg          string
		packagePaths stringSlice
	)
	flag.StringVar(&pkg, "pkg", "temporalcli", "Package name for generated code")
	flag.Var(&packagePaths, "package", "Nexus service import path to scan (can be specified multiple times)")
	flag.Parse()

	if len(packagePaths) == 0 {
		return fmt.Errorf("-package flag is required")
	}

	ops, err := systemnexusgen.Parse(packagePaths...)
	if err != nil {
		return fmt.Errorf("failed parsing nexus packages: %w", err)
	}

	b, err := systemnexusgen.GenerateCode(pkg, ops)
	if err != nil {
		return fmt.Errorf("failed generating code: %w", err)
	}

	if _, err := os.Stdout.Write(b); err != nil {
		return fmt.Errorf("failed writing output: %w", err)
	}
	return nil
}
