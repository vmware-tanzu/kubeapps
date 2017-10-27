package prototype

import (
	"fmt"
	"strings"
)

const (
	delimiter = "\x00"
)

type index struct {
	prototypes map[string]*SpecificationSchema
}

func (idx *index) List() (SpecificationSchemas, error) {
	prototypes := []*SpecificationSchema{}
	for _, prototype := range idx.prototypes {
		prototypes = append(prototypes, prototype)
	}
	return prototypes, nil
}

func (idx *index) SearchNames(query string, opts SearchOptions) (SpecificationSchemas, error) {
	// TODO(hausdorff): This is the world's worst search algorithm. Improve it at
	// some point.

	prototypes := []*SpecificationSchema{}

	for name, prototype := range idx.prototypes {
		isSearchResult := false
		switch opts {
		case Prefix:
			isSearchResult = strings.HasPrefix(name, query)
		case Suffix:
			isSearchResult = strings.HasSuffix(name, query)
		case Substring:
			isSearchResult = strings.Contains(name, query)
		default:
			return nil, fmt.Errorf("Unrecognized search option '%d'", opts)
		}

		if isSearchResult {
			prototypes = append(prototypes, prototype)
		}
	}

	return prototypes, nil
}
