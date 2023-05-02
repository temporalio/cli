package namespace

import (
	"fmt"
	"reflect"

	"github.com/temporalio/cli/common"
	"github.com/temporalio/tctl-kit/pkg/output"
	"github.com/urfave/cli/v2"
)

var (
	createNamespaceFlags = []cli.Flag{
		&cli.StringFlag{
			Name:     common.FlagDescription,
			Usage:    "Namespace description",
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagOwnerEmail,
			Usage:    "Owner email",
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagRetention,
			Usage:    "Workflow Execution retention",
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagActiveCluster,
			Usage:    "Active cluster name",
			Category: common.CategoryMain,
		},
		&cli.StringSliceFlag{
			Name:     common.FlagCluster,
			Usage:    "Cluster name",
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagIsGlobalNamespace,
			Usage:    "Flag to indicate whether namespace is a global namespace",
			Category: common.CategoryMain,
		},
		&cli.StringSliceFlag{
			Name:     common.FlagNamespaceData,
			Usage:    "Namespace data in a format key=value",
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagHistoryArchivalState,
			Usage:    "Flag to set history archival state, valid values are \"disabled\" and \"enabled\"",
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagHistoryArchivalURI,
			Usage:    "Optionally specify history archival URI (cannot be changed after first time archival is enabled)",
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagVisibilityArchivalState,
			Usage:    "Flag to set visibility archival state, valid values are \"disabled\" and \"enabled\"",
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagVisibilityArchivalURI,
			Usage:    "Optionally specify visibility archival URI (cannot be changed after first time archival is enabled)",
			Category: common.CategoryMain,
		},
	}

	updateNamespaceFlags = []cli.Flag{
		&cli.StringFlag{
			Name:     common.FlagDescription,
			Usage:    "Namespace description",
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagOwnerEmail,
			Usage:    "Owner email",
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagRetention,
			Usage:    "Workflow Execution retention",
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagActiveCluster,
			Usage:    "Active cluster name",
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagCluster,
			Usage:    "Cluster name",
			Category: common.CategoryMain,
		},
		&cli.StringSliceFlag{
			Name:     common.FlagNamespaceData,
			Usage:    "Namespace data in a format key=value",
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagHistoryArchivalState,
			Usage:    "Flag to set history archival state, valid values are \"disabled\" and \"enabled\"",
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagHistoryArchivalURI,
			Usage:    "Optionally specify history archival URI (cannot be changed after first time archival is enabled)",
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagVisibilityArchivalState,
			Usage:    "Flag to set visibility archival state, valid values are \"disabled\" and \"enabled\"",
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagVisibilityArchivalURI,
			Usage:    "Optionally specify visibility archival URI (cannot be changed after first time archival is enabled)",
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     common.FlagReason,
			Usage:    "Reason for the operation",
			Category: common.CategoryMain,
		},
		&cli.BoolFlag{
			Name:     common.FlagPromoteNamespace,
			Usage:    "Promote local namespace to global namespace",
			Category: common.CategoryMain,
		},
		&cli.StringFlag{
			Name:     output.FlagOutput,
			Aliases:  common.FlagOutputAlias,
			Usage:    output.UsageText,
			Value:    string(output.Table),
			Category: common.CategoryDisplay,
		},
		&cli.BoolFlag{
			Name:     common.FlagVerbose,
			Aliases:  common.FlagVerboseAlias,
			Usage:    "Print applied namespace changes",
			Category: common.CategoryDisplay,
		},
	}

	describeNamespaceFlags = []cli.Flag{
		&cli.StringFlag{
			Name:     common.FlagNamespaceID,
			Usage:    "Namespace Id",
			Category: common.CategoryMain,
		},
	}
)

type mutation struct {
	Field  string
	Before interface{}
	After  interface{}
}

func compareStructs(a, b interface{}) []mutation {
	var mutations []mutation
	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)

	// Dereference pointers if needed
	if va.Kind() == reflect.Ptr {
		va = va.Elem()
	}
	if vb.Kind() == reflect.Ptr {
		vb = vb.Elem()
	}

	// Check if both values are structs
	if va.Kind() != reflect.Struct || vb.Kind() != reflect.Struct {
		panic("Input values must be structs or pointers to structs")
	}

	compareStructsRecursively("", va, vb, &mutations)
	return mutations
}

func compareStructsRecursively(prefix string, va, vb reflect.Value, mutations *[]mutation) {
	for i := 0; i < va.NumField(); i++ {
		fieldA := va.Field(i)
		fieldB := vb.Field(i)
		fieldName := va.Type().Field(i).Name

		// Build the field path
		if prefix != "" {
			fieldName = fmt.Sprintf("%s.%s", prefix, fieldName)
		}

		// Dereference pointers if needed
		if fieldA.Kind() == reflect.Ptr {
			fieldA = fieldA.Elem()
		}
		if fieldB.Kind() == reflect.Ptr {
			fieldB = fieldB.Elem()
		}

		if fieldA.Kind() == reflect.Struct && fieldB.Kind() == reflect.Struct {
			// Recursively compare nested structures
			compareStructsRecursively(fieldName, fieldA, fieldB, mutations)
		} else {
			if !reflect.DeepEqual(fieldA.Interface(), fieldB.Interface()) {
				*mutations = append(*mutations, mutation{
					Field:  fieldName,
					Before: fieldA.Interface(),
					After:  fieldB.Interface(),
				})
			}
		}
	}
}
