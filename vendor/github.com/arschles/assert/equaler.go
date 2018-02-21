package assert

// Equaler determines if a type is equal to any other type that conforms to
// Equaler. All types passed to assert.Equal are checked to see if they
// conform to this interface, and if they do, their Equal function is called
// to check for their equality. This call is made instead of the call to
// reflect.DeepEqual
type Equaler interface {
	Equal(Equaler) bool
}
