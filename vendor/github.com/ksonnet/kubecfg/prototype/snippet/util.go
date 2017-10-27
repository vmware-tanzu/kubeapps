package snippet

func compareByIndex(a placeholder, b placeholder) int {
	if a.index == b.index {
		return 0
	} else if a.isFinalTabstop() {
		return 1
	} else if b.isFinalTabstop() {
		return -1
	} else if a.index < b.index {
		return -1
	} else if a.index > b.index {
		return 1
	}
	return 0
}

func walk(ms *markers, visitor func(m marker) bool) {
	stack := make(markers, len(*ms))
	copy(stack, *ms)

	for len(stack) > 0 {
		// NOTE: Declare `m` separately so that we can use the `=` operator
		// (rather than `:=`) to make it clear that we're not shadowing `stack`.
		var m marker
		m, stack = stack[0], stack[1:]
		recurse := visitor(m)
		if !recurse {
			break
		}
		stack = append(*m.children(), stack...)
	}
}

// * fill in default for empty placeHolders
// * compact sibling Text markers
func walkDefaults(ms *markers, placeholderDefaultValues map[int]*markers) {

	for i := 0; i < len(*ms); i++ {
		thisMarker := (*ms)[i]

		switch thisMarker.(type) {
		case *placeholder:
			{
				pl := thisMarker.(*placeholder)
				// fill in default values for repeated placeholders
				// like `${1:foo}and$1` becomes ${1:foo}and${1:foo}
				if defaultVal, ok := placeholderDefaultValues[pl.index]; !ok {
					placeholderDefaultValues[pl.index] = pl._children
					walkDefaults(pl._children, placeholderDefaultValues)

				} else if len(*pl._children) == 0 {
					// copy children from first placeholder definition, no need to
					// recurse on them because they have been visited already
					children := make(markers, len(*defaultVal))
					pl._children = &children
					copy(*pl._children, *defaultVal)
				}
			}
		case *variable:
			{
				walkDefaults(thisMarker.children(), placeholderDefaultValues)
			}
		case *text:
			{
				if i <= 0 {
					continue
				}

				prev := (*ms)[i-1]
				switch prev.(type) {
				case *text:
					{
						(*ms)[i-1].(*text).data += (*ms)[i].(*text).data
						ms.delete(i)
						i--
					}
				}
			}
		}
	}
}
