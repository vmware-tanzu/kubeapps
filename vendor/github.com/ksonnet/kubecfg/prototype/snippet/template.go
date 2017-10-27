package snippet

type textmateSnippet struct {
	markerImpl
	_placeholders *[]*placeholder
}

func newTextmateSnippet(children *markers) *textmateSnippet {
	tms := &textmateSnippet{
		markerImpl: markerImpl{
			_children: children,
		},
		_placeholders: nil,
	}
	tms._children.setParents(tms)
	return tms
}

func (tms *textmateSnippet) placeholders() []*placeholder {
	if tms._placeholders == nil {
		// Fill in placeholders if they don't exist.
		tms._placeholders = &[]*placeholder{}
		walk(tms._children, func(candidate marker) bool {
			switch candidate.(type) {
			case *placeholder:
				{
					*tms._placeholders = append(*tms._placeholders, candidate.(*placeholder))
				}
			}
			return true
		})
	}
	return *tms._placeholders
}

func (tms *textmateSnippet) offset(m marker) int {
	pos := 0
	found := false
	walk(tms._children, func(candidate marker) bool {
		if candidate == m {
			found = true
			return false
		}
		pos += candidate.len()
		return true
	})

	if !found {
		return -1
	}
	return pos
}

func (tms *textmateSnippet) fullLen(m marker) int {
	ret := 0
	walk(&markers{m}, func(m marker) bool {
		ret += m.len()
		return true
	})
	return ret
}

func (tms *textmateSnippet) enclosingPlaceholders(ph placeholder) []*placeholder {
	ret := []*placeholder{}
	parent := ph._parent
	for parent != nil {
		switch parent.(type) {
		case *placeholder:
			{
				ret = append(ret, parent.(*placeholder))
			}
		}
		parent = parent.parent()
	}
	return ret
}

func (tms *textmateSnippet) text() string {
	return tms._children.String()
}

func (tms *textmateSnippet) Evaluate(values map[string]string) (string, error) {
	walk(tms.children(), func(candidate marker) bool {
		switch casted := candidate.(type) {
		case *variable:
			{
				if resolved, ok := values[casted.name]; ok {
					casted.resolvedValue = &resolved
				}
				if casted.isDefined() {
					// remove default value from resolved variable
					casted._children = &markers{}
				}
			}
		}
		return true
	})

	// TODO: Explicitly disallow tabstops and empty placeholders. Error out if
	// present.

	return tms.text(), nil
}

func (tms *textmateSnippet) ReplacePlaceholder(idx index, replaceWith *markers) {
	newChildren := make(markers, len(*replaceWith))
	copy(newChildren, *replaceWith)
	newChildren.delete(int(idx))
	tms._children = &newChildren
	tms._placeholders = nil
}
