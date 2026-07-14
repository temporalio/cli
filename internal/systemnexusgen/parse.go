// Package systemnexusgen reads Temporal nexus service packages and generates
// the systemNexusOps map for the CLI.
package systemnexusgen

import (
	"fmt"
	"go/types"
	"sort"

	"golang.org/x/tools/go/packages"
)

// operationReferenceType is the fully qualified name of the generic nexus
// operation reference type whose type arguments are the request/response protos.
const operationReferenceType = "github.com/nexus-rpc/sdk-go/nexus.OperationReference"

// ProtoType identifies a proto message type by its package and type name.
type ProtoType struct {
	PkgPath string
	PkgName string
	Name    string
}

// Operation describes a single system Nexus operation found on a service struct.
type Operation struct {
	NexusPkgPath   string // import path of the package declaring the service struct
	NexusPkgName   string // package name of that package
	ServiceVarName string // e.g. TemporalAPIWorkflowserviceV1WorkflowService
	OpFieldName    string // e.g. SignalWithStartWorkflowExecution
	Request        ProtoType
	Response       ProtoType
}

// Parse loads the given import paths and returns all system Nexus operations
// found on their service structs, sorted deterministically.
func Parse(importPaths ...string) ([]Operation, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo |
			packages.NeedImports | packages.NeedDeps,
	}
	pkgs, err := packages.Load(cfg, importPaths...)
	if err != nil {
		return nil, fmt.Errorf("loading packages: %w", err)
	}

	var ops []Operation
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			return nil, fmt.Errorf("package %s has errors: %v", pkg.PkgPath, pkg.Errors)
		}
		ops = append(ops, operationsFromPackage(pkg)...)
	}

	sort.Slice(ops, func(i, j int) bool {
		if ops[i].NexusPkgPath != ops[j].NexusPkgPath {
			return ops[i].NexusPkgPath < ops[j].NexusPkgPath
		}
		if ops[i].ServiceVarName != ops[j].ServiceVarName {
			return ops[i].ServiceVarName < ops[j].ServiceVarName
		}
		return ops[i].OpFieldName < ops[j].OpFieldName
	})
	return ops, nil
}

func operationsFromPackage(pkg *packages.Package) []Operation {
	var ops []Operation
	scope := pkg.Types.Scope()
	for _, name := range scope.Names() {
		v, ok := scope.Lookup(name).(*types.Var)
		if !ok {
			continue
		}
		strct, ok := v.Type().Underlying().(*types.Struct)
		if !ok || !isServiceStruct(strct) {
			continue
		}
		for field := range strct.Fields() {
			req, resp, ok := operationRefTypeArgs(field.Type())
			if !ok {
				continue
			}
			ops = append(ops, Operation{
				NexusPkgPath:   pkg.PkgPath,
				NexusPkgName:   pkg.Name,
				ServiceVarName: v.Name(),
				OpFieldName:    field.Name(),
				Request:        req,
				Response:       resp,
			})
		}
	}
	return ops
}

// isServiceStruct reports whether strct looks like a nexus service struct:
// it has a `ServiceName string` field and at least one OperationReference field.
func isServiceStruct(strct *types.Struct) bool {
	hasServiceName, hasOperation := false, false
	for f := range strct.Fields() {
		if f.Name() == "ServiceName" {
			if b, ok := f.Type().(*types.Basic); ok && b.Kind() == types.String {
				hasServiceName = true
			}
		}
		if _, _, ok := operationRefTypeArgs(f.Type()); ok {
			hasOperation = true
		}
	}
	return hasServiceName && hasOperation
}

// operationRefTypeArgs returns the request/response proto types when t is an
// instantiated nexus.OperationReference[Req, Resp].
func operationRefTypeArgs(t types.Type) (ProtoType, ProtoType, bool) {
	named, ok := t.(*types.Named)
	if !ok {
		return ProtoType{}, ProtoType{}, false
	}
	obj := named.Obj()
	if obj.Pkg() == nil || obj.Pkg().Path()+"."+obj.Name() != operationReferenceType {
		return ProtoType{}, ProtoType{}, false
	}
	args := named.TypeArgs()
	if args == nil || args.Len() != 2 {
		return ProtoType{}, ProtoType{}, false
	}
	req, ok := protoTypeFrom(args.At(0))
	if !ok {
		return ProtoType{}, ProtoType{}, false
	}
	resp, ok := protoTypeFrom(args.At(1))
	if !ok {
		return ProtoType{}, ProtoType{}, false
	}
	return req, resp, true
}

func protoTypeFrom(t types.Type) (ProtoType, bool) {
	named, ok := t.(*types.Named)
	if !ok {
		return ProtoType{}, false
	}
	obj := named.Obj()
	if obj.Pkg() == nil {
		return ProtoType{}, false
	}
	return ProtoType{
		PkgPath: obj.Pkg().Path(),
		PkgName: obj.Pkg().Name(),
		Name:    obj.Name(),
	}, true
}
