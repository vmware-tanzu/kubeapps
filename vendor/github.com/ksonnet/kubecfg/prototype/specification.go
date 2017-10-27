package prototype

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

//
// NOTE: These members would ordinarily be private and exposed by interfaces,
// but because Go requires public structs for un/marshalling, it is more
// convenient to simply expose all of them.
//

// SpecificationSchema is the JSON-serializable representation of a prototype
// specification.
type SpecificationSchema struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`

	// Unique identifier of the mixin library. The most reliable way to make a
	// name unique is to embed a domain you own into the name, as is commonly done
	// in the Java community.
	Name     string        `json:"name"`
	Params   ParamSchemas  `json:"params"`
	Template SnippetSchema `json:"template"`
}

// SpecificationSchemas is a slice of pointer to `SpecificationSchema`.
type SpecificationSchemas []*SpecificationSchema

func (ss SpecificationSchemas) String() string {
	//
	// We want output that's lined up, like:
	//
	//   io.whatever.pkg.foo    Foo's main template    [jsonnet, yaml]
	//   io.whatever.pkg.foobar Foobar's main template [jsonnet, yaml, json]
	//
	// To accomplish this we find (1) the longest prototype name, and (2) the
	// longest description, so that we can properly pad the output.
	//

	maxNameLen := 0
	for _, proto := range ss {
		if l := len(proto.Name); l > maxNameLen {
			maxNameLen = l
		}
	}

	maxNameDescLen := 0
	for _, proto := range ss {
		nameDescLen := maxNameLen + 1 + len(proto.Template.ShortDescription)
		if nameDescLen > maxNameDescLen {
			maxNameDescLen = nameDescLen
		}
	}

	lines := []string{}
	for _, proto := range ss {
		// NOTE: If we don't add 1 below, the longest name will look like :
		// `io.whatever.pkg.fooDescription is here.`
		nameSpace := strings.Repeat(" ", maxNameLen-len(proto.Name)+1)
		descSpace := strings.Repeat(" ", maxNameDescLen-maxNameLen-len(proto.Template.ShortDescription)+2)

		avail := fmt.Sprintf("%s", proto.Template.AvailableTemplates())

		lines = append(lines, proto.Name+nameSpace+proto.Template.ShortDescription+descSpace+avail+"\n")
	}

	sort.Slice(lines, func(i, j int) bool { return lines[i] < lines[j] })

	return strings.Join(lines, "")
}

// RequiredParams retrieves all parameters that are required by a prototype.
func (s *SpecificationSchema) RequiredParams() ParamSchemas {
	reqd := ParamSchemas{}
	for _, p := range s.Params {
		if p.Default == nil {
			reqd = append(reqd, p)
		}
	}

	return reqd
}

// OptionalParams retrieves all parameters that can optionally be provided to a
// prototype.
func (s *SpecificationSchema) OptionalParams() ParamSchemas {
	opt := ParamSchemas{}
	for _, p := range s.Params {
		if p.Default != nil {
			opt = append(opt, p)
		}
	}

	return opt
}

// TemplateType represents the possible type of a prototype.
type TemplateType string

const (
	// YAML represents a prototype written in YAML.
	YAML TemplateType = "yaml"

	// JSON represents a prototype written in JSON.
	JSON TemplateType = "json"

	// Jsonnet represents a prototype written in Jsonnet.
	Jsonnet TemplateType = "jsonnet"
)

// ParseTemplateType attempts to parse a string as a `TemplateType`.
func ParseTemplateType(t string) (TemplateType, error) {
	switch strings.ToLower(t) {
	case "yaml":
		return YAML, nil
	case "json":
		return JSON, nil
	case "jsonnet":
		return Jsonnet, nil
	default:
		return "", fmt.Errorf("Unrecognized template type '%s'; must be one of: [yaml, json, jsonnet]", t)
	}
}

// SnippetSchema is the JSON-serializable representation of the TextMate snippet
// specification, as implemented by the Language Server Protocol.
type SnippetSchema struct {
	Prefix string `json:"prefix"`

	// Description describes what the prototype does.
	Description string `json:"description"`

	// ShortDescription briefly describes what the prototype does.
	ShortDescription string `json:"shortDescription"`

	// Various body types of the prototype. Follows the TextMate snippets syntax,
	// with several features disallowed. At least one of these is required to be
	// filled out.
	JSONBody    []string `json:"jsonBody"`
	YAMLBody    []string `json:"yamlBody"`
	JsonnetBody []string `json:"jsonnetBody"`
}

// Body attempts to retrieve the template body associated with some
// type `t`.
func (schema *SnippetSchema) Body(t TemplateType) (template []string, err error) {
	switch t {
	case YAML:
		template = schema.YAMLBody
	case JSON:
		template = schema.JSONBody
	case Jsonnet:
		template = schema.JsonnetBody
	default:
		return nil, fmt.Errorf("Unrecognized template type '%s'; must be one of: [yaml, json, jsonnet]", t)
	}

	if len(template) == 0 {
		available := schema.AvailableTemplates()
		err = fmt.Errorf("Template does not have a template for type '%s'. Available types: %s", t, available)
	}

	return
}

// AvailableTemplates returns the list of available `TemplateType`s this
// prototype implements.
func (schema *SnippetSchema) AvailableTemplates() (ts []TemplateType) {
	if len(schema.YAMLBody) != 0 {
		ts = append(ts, YAML)
	}

	if len(schema.JSONBody) != 0 {
		ts = append(ts, JSON)
	}

	if len(schema.JsonnetBody) != 0 {
		ts = append(ts, Jsonnet)
	}

	return
}

// ParamType represents a type constraint for a prototype parameter (e.g., it
// must be a number).
type ParamType string

const (
	// Number represents a prototype parameter that must be a number.
	Number ParamType = "number"

	// String represents a prototype parameter that must be a string.
	String ParamType = "string"

	// NumberOrString represents a prototype parameter that must be either a
	// number or a string.
	NumberOrString ParamType = "numberOrString"

	// Object represents a prototype parameter that must be an object.
	Object ParamType = "object"

	// Array represents a prototype parameter that must be a array.
	Array ParamType = "array"
)

func (pt ParamType) String() string {
	switch pt {
	case Number:
		return "number"
	case String:
		return "string"
	case NumberOrString:
		return "number-or-string"
	case Object:
		return "object"
	case Array:
		return "array"
	default:
		return "unknown"
	}
}

// ParamSchema is the JSON-serializable representation of a parameter provided
// to a prototype.
type ParamSchema struct {
	Name        string    `json:"name"`
	Alias       *string   `json:"alias"` // Optional.
	Description string    `json:"description"`
	Default     *string   `json:"default"` // `nil` only if the parameter is optional.
	Type        ParamType `json:"type"`
}

// Quote will parse a prototype parameter and quote it appropriately, so that it
// shows up correctly in Jsonnet source code. For example, `--image nginx` would
// likely need to show up as `"nginx"` in Jsonnet source.
func (ps *ParamSchema) Quote(value string) (string, error) {
	switch ps.Type {
	case Number:
		_, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return "", fmt.Errorf("Could not convert parameter '%s' to a number", ps.Name)
		}
		return value, nil
	case String:
		return fmt.Sprintf("\"%s\"", value), nil
	case NumberOrString:
		_, err := strconv.ParseFloat(value, 64)
		if err == nil {
			return value, nil
		}
		return fmt.Sprintf("\"%s\"", value), nil
	case Array, Object:
		return value, nil
	default:
		return "", fmt.Errorf("Unknown param type for param '%s'", ps.Name)
	}
}

// RequiredParam constructs a required parameter, i.e., a parameter that is
// meant to be required by some prototype, somewhere.
func RequiredParam(name, alias, description string, t ParamType) *ParamSchema {
	return &ParamSchema{
		Name:        name,
		Alias:       &alias,
		Description: description,
		Default:     nil,
		Type:        t,
	}
}

// OptionalParam constructs an optional parameter, i.e., a parameter that is
// meant to be optionally provided to some prototype, somewhere.
func OptionalParam(name, alias, description, defaultVal string, t ParamType) *ParamSchema {
	return &ParamSchema{
		Name:        name,
		Alias:       &alias,
		Description: description,
		Default:     &defaultVal,
		Type:        t,
	}
}

// ParamSchemas is a slice of `ParamSchema`
type ParamSchemas []*ParamSchema

// PrettyString creates a prettified string representing a collection of
// parameters.
func (ps ParamSchemas) PrettyString(prefix string) string {
	if len(ps) == 0 {
		return "  [none]"
	}

	flags := []string{}
	for _, p := range ps {
		alias := p.Name
		if p.Alias != nil {
			alias = *p.Alias
		}
		flags = append(flags, fmt.Sprintf("--%s=<%s>", p.Name, alias))
	}

	max := 0
	for _, flag := range flags {
		if flagLen := len(flag); max < flagLen {
			max = flagLen
		}
	}

	prettyFlags := []string{}
	for i := range flags {
		p := ps[i]
		flag := flags[i]

		var info string
		if p.Default != nil {
			info = fmt.Sprintf(" [default: %s, type: %s]", *p.Default, p.Type.String())
		} else {
			info = fmt.Sprintf(" [type: %s]", p.Type.String())
		}

		// NOTE: If we don't add 1 here, the longest line will look like:
		// `--flag=<flag>Description is here.`
		space := strings.Repeat(" ", max-len(flag)+1)
		pretty := fmt.Sprintf(prefix + flag + space + p.Description + info)
		prettyFlags = append(prettyFlags, pretty)
	}

	return strings.Join(prettyFlags, "\n")
}
