package jsonnet

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

// Demo for the README.
func Test_Demo(t *testing.T) {
	vm := Make()
	vm.ExtVar("color", "purple")

	x, err := vm.EvaluateSnippet(`Test_Demo`, `"dark " + std.extVar("color")`)
	if err != nil {
		panic(err)
	}
	if x != "\"dark purple\"\n" {
		panic("fail: we got " + x)
	}
	vm.Destroy()
}

// importFunc returns a couple of hardwired responses.
func importFunc(base, rel string) (result string, path string, err error) {
	if rel == "alien.conf" {
		return `{ type: "alien", origin: "Ork", name: "Mork" }`, "alien.conf", nil
	}
	if rel == "human.conf" {
		return `{ type: "human", origin: "Earth", name: "Mendy" }`, "human.conf", nil
	}
	return "", "", errors.New(fmt.Sprintf("Cannot import %q", rel))
}

// check there is no err, and a == b.
func check(t *testing.T, err error, a, b string) {
	if err != nil {
		t.Errorf("got error: %q", err.Error())
	}
	if a != b {
		t.Errorf("got %q but wanted %q", a, b)
	}
}

func Test_Simple(t *testing.T) {

	// Each time there's a new version, this will force an update to this code.
	check(t, nil, Version(), `v0.9.4`)

	vm := Make()
	defer vm.Destroy()
	vm.TlaVar("color", "purple")
	vm.TlaVar("size", "XXL")
	vm.TlaCode("gooselevel", "1234 * 10 + 5")
	vm.ExtVar("color", "purple")
	vm.ExtVar("size", "XXL")
	vm.ExtCode("gooselevel", "1234 * 10 + 5")
	vm.ImportCallback(importFunc)

	x, err := vm.EvaluateSnippet(`test1`, `20 + 22`)
	check(t, err, x, `42`+"\n")
	x, err = vm.EvaluateSnippet(`test2`, `function(color, size, gooselevel) color`)
	check(t, err, x, `"purple"`+"\n")
	x, err = vm.EvaluateSnippet(`test2`, `std.extVar("color")`)
	check(t, err, x, `"purple"`+"\n")
	vm.StringOutput(true)
	x, err = vm.EvaluateSnippet(`test2`, `"whee"`)
	check(t, err, x, `whee`+"\n")
	vm.StringOutput(false)
	x, err = vm.EvaluateSnippet(`test3`, `
    local a = import "alien.conf";
    local b = import "human.conf";
    a.name + b.name
    `)
	check(t, err, x, `"MorkMendy"`+"\n")
	x, err = vm.EvaluateSnippet(`test4`, `
    local a = import "alien.conf";
    local b = a { type: "fictitious" };
    b.type + b.name
    `)
	check(t, err, x, `"fictitiousMork"`+"\n")
}

func Test_FileScript(t *testing.T) {
	vm := Make()
	defer vm.Destroy()
	x, err := vm.EvaluateFile("test2.j")
	check(t, err, x, `{
   "awk": "/usr/bin/awk",
   "shell": "/bin/csh"
}
`)
}

func Test_Misc(t *testing.T) {
	vm := Make()
	defer vm.Destroy()

	vm.MaxStack(10)
	vm.MaxTrace(10)
	vm.GcMinObjects(10)
	vm.GcGrowthTrigger(2.0)

	x, err := vm.EvaluateSnippet("Misc", `
    local a = import "test2.j";
    a.awk + a.shell`)
	check(t, err, x, `"/usr/bin/awk/bin/csh"`+"\n")
}

func Test_FormatFile(t *testing.T) {
	f, err := ioutil.TempFile("", "jsonnet-fmt-test")
	if err != nil {
		t.Fatal(err)
	}
	filename := f.Name()
	defer func() {
		f.Close()
		os.Remove(filename)
	}()

	data := `{
    "quoted": "keys",
    "notevaluated": 20 + 22,
    "trailing": "comma",}
`
	if err := ioutil.WriteFile(filename, []byte(data), 0644); err != nil {
		t.Fatalf("WriteFile %s: %v", filename, err)
	}

	vm := Make()
	defer vm.Destroy()
	result, err := vm.FormatFile(filename)

	check(t, err, result, `{
    quoted: "keys",
    notevaluated: 20 + 22,
    trailing: "comma" }
`)
}

func Test_FormatSnippet(t *testing.T) {
	data := `{
    "quoted": "keys",
    "notevaluated": 20 + 22,
    "trailing": "comma",}
`

	vm := Make()
	defer vm.Destroy()
	result, err := vm.FormatSnippet("testfoo", data)

	check(t, err, result, `{
    quoted: "keys",
    notevaluated: 20 + 22,
    trailing: "comma" }
`)
}

func Test_FormatIndent(t *testing.T) {
	data := `{
  "quoted": "keys",
 "notevaluated": 20 + 22,
   "trailing": "comma",}
`

	vm := Make()
	defer vm.Destroy()
	vm.FormatIndent(1)
	result, err := vm.FormatSnippet("testfoo", data)

	check(t, err, result, `{
 quoted: "keys",
 notevaluated: 20 + 22,
 trailing: "comma" }
`)
}

func TestJsonString(t *testing.T) {
	vm := Make()
	defer vm.Destroy()

	v := vm.NewString("foo")

	if v2, ok := v.ExtractString(); !ok {
		t.Errorf("ExtractString() returned !ok")
	} else if v2 != "foo" {
		t.Errorf("Incorrect extracted string: %s", v2)
	}

	if _, ok := v.ExtractNumber(); ok {
		t.Errorf("ExtractNumber() returned ok")
	}

	if _, ok := v.ExtractBool(); ok {
		t.Errorf("ExtractBool() returned ok")
	}

	if ok := v.ExtractNull(); ok {
		t.Errorf("ExtractNull() returned ok")
	}
}

func TestJsonNumber(t *testing.T) {
	vm := Make()
	defer vm.Destroy()

	v := vm.NewNumber(42)

	if _, ok := v.ExtractString(); ok {
		t.Errorf("ExtractString() returned ok")
	}

	if v2, ok := v.ExtractNumber(); !ok {
		t.Errorf("ExtractNumber() returned !ok")
	} else if v2 != 42 {
		t.Errorf("ExtractNumber() returned unexpected value: %v", v2)
	}

	if _, ok := v.ExtractBool(); ok {
		t.Errorf("ExtractBool() returned ok")
	}

	if ok := v.ExtractNull(); ok {
		t.Errorf("ExtractNull() returned ok")
	}
}

func TestJsonArray(t *testing.T) {
	vm := Make()
	defer vm.Destroy()

	a := vm.NewArray()
	a.ArrayAppend(vm.NewString("foo"))
	a.ArrayAppend(vm.NewNull())
	a.ArrayAppend(vm.NewNumber(3.14))

	// Can't actually inspect array elements with this version of
	// jsonnet ...
}

func TestJsonObject(t *testing.T) {
	vm := Make()
	defer vm.Destroy()

	a := vm.NewObject()
	a.ObjectAppend("foo", vm.NewString("foo"))
	a.ObjectAppend("bar", vm.NewNull())
	a.ObjectAppend("baz", vm.NewNumber(3.14))

	// Can't actually inspect array elements with this version of
	// jsonnet ...
}

func TestNativeRaw(t *testing.T) {
	vm := Make()
	defer vm.Destroy()

	vm.NativeCallbackRaw("myadd", []string{"a", "b"}, func(args ...*JsonValue) (*JsonValue, error) {
		if len(args) != 2 {
			return nil, errors.New("wrong number of args")
		}

		a, ok := args[0].ExtractNumber()
		if !ok {
			return nil, errors.New("a is not a number")
		}
		b, ok := args[1].ExtractNumber()
		if !ok {
			return nil, errors.New("b is not a number")
		}
		return vm.NewNumber(a + b), nil
	})

	x, err := vm.EvaluateSnippet("NativeRaw", `std.native("myadd")(3, 4)`)
	check(t, err, x, "7\n")

	x, err = vm.EvaluateSnippet("NativeRaw", `std.native("myadd")(42)`)
	if err == nil {
		t.Errorf("Wrong number of args failed to produce error")
	}

	x, err = vm.EvaluateSnippet("NativeRaw", `std.native("myadd")(3, "foo")`)
	if err == nil {
		t.Errorf("Go error was not translated into jsonnet error")
	} else if !strings.Contains(err.Error(), "b is not a number") {
		t.Errorf("Wrong jsonnet error: %v", err)
	}
}

func TestNative(t *testing.T) {
	vm := Make()
	defer vm.Destroy()

	vm.NativeCallback("myadd", []string{"a", "b"}, func(a, b float64) (float64, error) {
		return a + b, nil
	})
	vm.NativeCallback("fail", []string{}, func() (interface{}, error) {
		return nil, fmt.Errorf("this is an error")
	})

	x, err := vm.EvaluateSnippet("Native", `std.native("myadd")(3, 4)`)
	check(t, err, x, "7\n")

	x, err = vm.EvaluateSnippet("Native", `std.native("myadd")(42)`)
	if err == nil {
		t.Errorf("Wrong number of args failed to produce error")
	}

	x, err = vm.EvaluateSnippet("Native", `std.native("myadd")(3, "foo")`)
	if err == nil {
		t.Errorf("Wrong arg types failed to produce an error")
	}

	x, err = vm.EvaluateSnippet("Native", `std.native("fail")()`)
	if err == nil {
		t.Errorf("Go error was not propagated")
	} else if !strings.Contains(err.Error(), "this is an error") {
		t.Errorf("Go error text was not propagated")
	}
}
