// Copyright 2017 The kubecfg authors
//
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package prototype

import (
	"errors"
	"sort"
	"strings"

	"github.com/google/go-jsonnet/ast"
	"github.com/google/go-jsonnet/parser"

	"github.com/ksonnet/kubecfg/prototype/snippet"
)

const (
	paramPrefix            = "param://"
	paramReplacementPrefix = "${"
	paramReplacementSuffix = "}"
)

// Parse rewrites the imports in a Jsonnet file before returning the parsed
// TextMate snippet.
func Parse(fn string, jsonnet string) (snippet.Template, error) {
	s, err := parse(fn, jsonnet)
	if err != nil {
		return nil, err
	}

	return snippet.Parse(s), nil
}

func parse(fn string, jsonnet string) (string, error) {
	tokens, err := parser.Lex(fn, jsonnet)
	if err != nil {
		return "", err
	}

	root, err := parser.Parse(tokens)
	if err != nil {
		return "", err
	}

	var imports []ast.Import

	// Gather all parameter imports
	err = visit(root, &imports)
	if err != nil {
		return "", err
	}

	// Replace all parameter imports
	return replace(jsonnet, imports), nil
}

// ---------------------------------------------------------------------------

func visit(node ast.Node, imports *[]ast.Import) error {
	switch n := node.(type) {
	case *ast.Import:
		// Add parameter-type imports to the list of replacements.
		if strings.HasPrefix(n.File, paramPrefix) {
			param := strings.TrimPrefix(n.File, paramPrefix)
			if len(param) < 1 {
				return errors.New("There must be a parameter following import param://")
			}
			*imports = append(*imports, *n)
		}
	case *ast.Apply:
		for _, arg := range n.Arguments {
			err := visit(arg, imports)
			if err != nil {
				return err
			}
		}
		return visit(n.Target, imports)
	case *ast.ApplyBrace:
		err := visit(n.Left, imports)
		if err != nil {
			return err
		}
		return visit(n.Right, imports)
	case *ast.Array:
		for _, element := range n.Elements {
			err := visit(element, imports)
			if err != nil {
				return err
			}
		}
	case *ast.ArrayComp:
		for _, spec := range n.Specs {
			err := visitCompSpec(spec, imports)
			if err != nil {
				return err
			}
		}
		return visit(n.Body, imports)
	case *ast.Assert:
		err := visit(n.Cond, imports)
		if err != nil {
			return err
		}
		err = visit(n.Message, imports)
		if err != nil {
			return err
		}
		return visit(n.Rest, imports)
	case *ast.Binary:
		err := visit(n.Left, imports)
		if err != nil {
			return err
		}
		return visit(n.Right, imports)
	case *ast.Conditional:
		err := visit(n.BranchFalse, imports)
		if err != nil {
			return err
		}
		err = visit(n.BranchTrue, imports)
		if err != nil {
			return err
		}
		return visit(n.Cond, imports)
	case *ast.Error:
		return visit(n.Expr, imports)
	case *ast.Function:
		return visit(n.Body, imports)
	case *ast.Index:
		err := visit(n.Target, imports)
		if err != nil {
			return err
		}
		return visit(n.Index, imports)
	case *ast.Slice:
		err := visit(n.Target, imports)
		if err != nil {
			return err
		}
		err = visit(n.BeginIndex, imports)
		if err != nil {
			return err
		}
		err = visit(n.EndIndex, imports)
		if err != nil {
			return err
		}
		return visit(n.Step, imports)
	case *ast.Local:
		for _, bind := range n.Binds {
			err := visitLocalBind(bind, imports)
			if err != nil {
				return err
			}
		}
		return visit(n.Body, imports)
	case *ast.Object:
		for _, field := range n.Fields {
			err := visitObjectField(field, imports)
			if err != nil {
				return err
			}
		}
	case *ast.DesugaredObject:
		for _, assert := range n.Asserts {
			err := visit(assert, imports)
			if err != nil {
				return err
			}
		}
		for _, field := range n.Fields {
			err := visitDesugaredObjectField(field, imports)
			if err != nil {
				return err
			}
		}
	case *ast.ObjectComp:
		for _, field := range n.Fields {
			err := visitObjectField(field, imports)
			if err != nil {
				return err
			}
		}
		for _, spec := range n.Specs {
			err := visitCompSpec(spec, imports)
			if err != nil {
				return err
			}
		}
	case *ast.ObjectComprehensionSimple:
		err := visit(n.Field, imports)
		if err != nil {
			return err
		}
		err = visit(n.Value, imports)
		if err != nil {
			return err
		}
		return visit(n.Array, imports)
	case *ast.SuperIndex:
		return visit(n.Index, imports)
	case *ast.InSuper:
		return visit(n.Index, imports)
	case *ast.Unary:
		return visit(n.Expr, imports)
	// The below nodes do not have any child nodes, but visit them anyway to
	// have the capability to error out on unsupported nodes that may later
	// be added to go-jsonnet.
	case *ast.ImportStr:
	case *ast.Dollar:
	case *ast.LiteralBoolean:
	case *ast.LiteralNull:
	case *ast.LiteralNumber:
	case *ast.LiteralString:
	case *ast.Self:
	case *ast.Var:
	case nil:
		return nil
	default:
		return errors.New("Unsupported ast.Node type found")
	}

	return nil
}

func visitCompSpec(node ast.CompSpec, imports *[]ast.Import) error {
	return visit(node.Expr, imports)
}

func visitObjectField(node ast.ObjectField, imports *[]ast.Import) error {
	err := visit(node.Expr1, imports)
	if err != nil {
		return err
	}
	err = visit(node.Expr2, imports)
	if err != nil {
		return err
	}
	return visit(node.Expr3, imports)
}

func visitDesugaredObjectField(node ast.DesugaredObjectField, imports *[]ast.Import) error {
	err := visit(node.Name, imports)
	if err != nil {
		return err
	}
	return visit(node.Body, imports)
}

func visitLocalBind(node ast.LocalBind, imports *[]ast.Import) error {
	return visit(node.Body, imports)
}

// ---------------------------------------------------------------------------

// replace converts all parameters in the passed Jsonnet of form
// `import 'param://port'` into `${port}`.
func replace(jsonnet string, imports []ast.Import) string {
	lines := strings.Split(jsonnet, "\n")

	// Imports must be sorted by reverse location to avoid indexing problems
	// during string replacement.
	sort.Slice(imports, func(i, j int) bool {
		if imports[i].Loc().End.Line == imports[j].Loc().End.Line {
			return imports[i].Loc().End.Column > imports[j].Loc().End.Column
		}
		return imports[i].Loc().End.Line > imports[j].Loc().End.Line
	})

	for _, im := range imports {
		param := paramReplacementPrefix + strings.TrimPrefix(im.File, paramPrefix) + paramReplacementSuffix

		lineStart := im.Loc().Begin.Line
		lineEnd := im.Loc().End.Line
		colStart := im.Loc().Begin.Column
		colEnd := im.Loc().End.Column

		// Case where import param is split over multiple strings.
		if lineEnd != lineStart {
			// Replace all intermediate lines with the empty string.
			for i := lineStart; i < lineEnd-1; i++ {
				lines[i] = ""
			}
			// Remove import param related logic from the last line.
			lines[lineEnd-1] = lines[lineEnd-1][colEnd:len(lines[lineEnd-1])]
			// Perform replacement in the first line of import param occurance.
			lines[lineStart-1] = lines[lineStart-1][:colStart-1] + param
		} else {
			line := lines[lineStart-1]
			lines[lineStart-1] = line[:colStart-1] + param + line[colEnd:len(line)]
		}
	}

	return strings.Join(lines, "\n")
}
