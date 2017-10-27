package prototype

import "encoding/json"

// Unmarshal takes the bytes of a JSON-encoded prototype specification, and
// deserializes them to a `SpecificationSchema`.
func Unmarshal(bytes []byte) (*SpecificationSchema, error) {
	var p SpecificationSchema
	err := json.Unmarshal(bytes, &p)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

// SearchOptions represents the type of prototype search to execute on an
// `Index`.
type SearchOptions int

const (
	// Prefix represents a search over prototype name prefixes.
	Prefix SearchOptions = iota

	// Suffix represents a search over prototype name suffices.
	Suffix

	// Substring represents a search over substrings of prototype names.
	Substring
)

// Index represents a queryable index of prototype specifications.
type Index interface {
	List() (SpecificationSchemas, error)
	SearchNames(query string, opts SearchOptions) (SpecificationSchemas, error)
}

// NewIndex constructs an index of prototype specifications from a list.
func NewIndex(prototypes []*SpecificationSchema) Index {
	idx := map[string]*SpecificationSchema{}

	for _, p := range defaultPrototypes {
		idx[p.Name] = p
	}

	for _, p := range prototypes {
		idx[p.Name] = p
	}

	return &index{
		prototypes: idx,
	}
}
