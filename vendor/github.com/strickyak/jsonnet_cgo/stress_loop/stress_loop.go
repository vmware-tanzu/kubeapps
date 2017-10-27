// A program to stress the JSonnet VM and its cgo wrappings.
// Run `top` in another shell, to see how much we leak memory.
package main

import (
	"flag"
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"

	J "github.com/strickyak/jsonnet_cgo"
)

type Unit struct{}

var L = flag.Int("l", 1000, "Repeat the big loop this many times")
var N = flag.Int("n", 100, "Repeat quick operations this many times.")
var P = flag.Int("p", 100, "How many goroutines running a JSonnet vm in parallel.")
var F = fmt.Sprintf
var Quiet = false

// Say something to the log, unless Quiet.
func Say(format string, args ...interface{}) {
	if !Quiet {
		log.Printf(format, args...)
	}
}

// Repeat the action so many times.
func Repeat(times int, action func()) {
	for i := 0; i < times; i++ {
		action()
	}
}

// ExercizeOneVM and then signal it is finished.
// It panics only if something goes really wrong.
func ExercizeOneVM(finished chan Unit) {
	vm := J.Make()
	vm.JpathAdd("/tmp")

	// concatStringsNative is an instance of NativeCallback.
	// It extracts string representations of string, number, bool, or null arguments.
	concatStringsNativeRaw := func(args ...*J.JsonValue) (result *J.JsonValue, err error) {
		z := ""
		for _, a := range args {
			z += F("%v", a.Extract())
		}
		return vm.NewString(z), nil
	}
	concat4Strings := func(a, b, c, d string) (string, error) { return a + b + c + d, nil }

	Repeat(*N, func() { vm.MaxStack(999) })
	Repeat(*N, func() { vm.MaxTrace(0) })
	Repeat(*N, func() { vm.GcMinObjects(10) })
	Repeat(*N, func() { vm.GcGrowthTrigger(2.0) })
	Repeat(*N, func() { vm.ExtVar("color", "purple") })
	Repeat(*N, func() { vm.TlaVar("shade", "dark") })

	// Concerning Native Callbacks,
	// See https://gist.github.com/sparkprime/5b2ab0a1b72beceab2cf5ea524db228c
	// and https://github.com/google/jsonnet/issues/108
	Repeat(*N, func() { vm.NativeCallbackRaw("Concat2", []string{"a", "b"}, concatStringsNativeRaw) })
	Repeat(*N, func() { vm.NativeCallbackRaw("Concat3", []string{"a", "b", "c"}, concatStringsNativeRaw) })
	Repeat(*N, func() { vm.NativeCallback("Concat4", []string{"a", "b", "c", "d"}, concat4Strings) })
	{
		got, err := vm.EvaluateSnippet(
			"SNIPPET-1",
			`function(shade) (std.native("Concat3")(shade, "-", std.extVar("color")))`)
		if err != nil {
			panic(err)
		}
		Say("1: %q", got)
		want := "\"dark-purple\"\n"
		if got != want {
			panic(F("got %q wanted %q", got, want))
		}
	}
	Repeat(*N, func() { vm.TlaCode("amount", "8 * 111") })
	{
		got, err := vm.EvaluateSnippet(
			"SNIPPET-2",
			`function(shade, amount) (amount * 10)`)
		if err != nil {
			panic(err)
		}
		Say("1b: %s", got)
		want := "8880\n"
		if got != want {
			panic(F("got %q wanted %q", got, want))
		}
	}
	Repeat(*N, func() { vm.StringOutput(true) })
	{
		// Provoke an error.  Pass 3 instead of 4 arguments to std.native("Concat4").
		_, err := vm.EvaluateSnippet(
			"SNIPPET-3",
			`function(shade, amount) (
				std.native("Concat4")(
					std.native("Concat2")(shade, "-"),
					"",
					std.extVar("color")))`)
		if !strings.HasPrefix(
			F("%v", err),
			"RUNTIME ERROR: Function parameter d not bound in call.") {
			panic(F("SNIPPET-3: Got wrong error string prefix: %q", err))
		}
	}
	{
		got, err := vm.EvaluateSnippet(
			"SNIPPET-4",
			`function(shade, amount) (
				std.native("Concat4")(
					std.native("Concat2")(shade, "-"),
					"",
					"",
					std.extVar("color")))`)
		if err != nil {
			panic(err)
		}
		Say("2: %s", got)
		want := "dark-purple\n"
		if got != want {
			panic(F("got %q wanted %q", got, want))
		}
	}
	Repeat(*N, func() {
		_, err := vm.FormatSnippet(
			"SNIPPET-5",
			`{ "hello" : "world" }`)
		if err != nil {
			panic(err)
		}
	})
	Repeat(*N, func() { _ = J.Version() })
	vm.Destroy()
	finished <- Unit{}
}

func main() {
	flag.Parse()
	startMain := time.Now()
	for loops := 1; loops <= *L; loops++ {
		startLoop := time.Now()

		finished := make(chan Unit, *P)
		for i := 0; i < *P; i++ {
			go ExercizeOneVM(finished) // Start them in parallel.
		}
		Say("Waiting...")
		for i := 0; i < *P; i++ {
			<-finished // Wait for all to finish.
		}

		duration := time.Now().Sub(startLoop)
		totalDuration := time.Now().Sub(startMain)
		log.Printf("Finished %d loops.  This loop %.3f sec.  Average %.3f sec per loop.",
			loops, duration.Seconds(), totalDuration.Seconds()/float64(loops))

		runtime.GC() // Don't let garbage accumulate, unless it's leaked.

		// Only be verbose on the first pass.
		Quiet = true
	}
}
