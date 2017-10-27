/*
Command line tool to try evaluating JSonnet.

Demos:
  echo "{ a: 1, b: 2 }"  | go run jsonnet_main/main.go /dev/stdin
  go run jsonnet_main/main.go test1.j
  go run jsonnet_main/main.go test2.j
  echo 'std.extVar("a") + "bar"' | go run jsonnet_main/main.go /dev/stdin a=foo
*/
package main

import "github.com/strickyak/jsonnet_cgo"

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

var stringOutput = flag.Bool(
	"string_output", false,
	"If set, will expect a string and output it verbatim")

func importFunc(base, rel string) (result string, path string, err error) {
	filename := filepath.Join(base, rel)
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", "", err
	}
	return string(contents), filename, nil
}

func main() {
	flag.Parse()
	vm := jsonnet.Make()
	vm.ImportCallback(importFunc)

	if stringOutput != nil {
		vm.StringOutput(*stringOutput)
	}

	args := flag.Args()
	if len(args) < 1 {
		log.Fatal("Usage:  jsonnet_main filename key1=val1 key2=val2...")
	}

	for i := 1; i < len(args); i++ {
		kv := strings.SplitN(args[i], "=", 2)
		if len(kv) != 2 {
			log.Fatalf("Error in jsonnet_main: Expected arg to be 'key=value': %q", args[i])
		}
		vm.ExtVar(kv[0], kv[1])
	}

	z, err := vm.EvaluateFile(args[0])
	if err != nil {
		log.Fatalf("Error in jsonnet_main: %s", err)
	}
	fmt.Print(z)

	vm.Destroy()
}
