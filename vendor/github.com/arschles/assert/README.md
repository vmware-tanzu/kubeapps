# assert

[![Build Status](https://travis-ci.org/arschles/assert.svg?branch=master)](https://travis-ci.org/arschles/assert)
[![GoDoc](https://godoc.org/github.com/arschles/assert?status.svg)](https://godoc.org/github.com/arschles/assert)
[![Go Report Card](http://goreportcard.com/badge/arschles/assert)](http://goreportcard.com/report/arschles/assert)

`assert` is [Go](http://golang.org/) package that provides convenience methods
for writing assertions in [standard Go tests](http://godoc.org/testing).

You can write this test with `assert`:

```go
func TestSomething(t *testing.T) {
  i, err := doSomething()
  assert.NoErr(t, err)
  assert.Equal(t, i, 123, "returned integer")
}
```

Instead of writing this test with only the standard `testing` library:

```go
func TestSomething(t *testing.T) {
  i, err := doSomething()
  if err != nil {
    t.Fatalf("error encountered: %s", err)
  }
  if i != 123 {
    t.Fatalf("returned integer was %d, not %d", i, 123)
  }
}
```
