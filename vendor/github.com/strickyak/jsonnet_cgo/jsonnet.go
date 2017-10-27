/*
jsonnet is a simple Go wrapper for the JSonnet VM.

See http://jsonnet.org/
*/
package jsonnet

// By Henry Strickland <@yak.net:strick>
// Made self-contained by Marko Mikulicic <mkm@bitnami.com>

/*
#include <memory.h>
#include <string.h>
#include <stdio.h>
#include <stdlib.h>
#include <libjsonnet.h>

char *CallImport_cgo(void *ctx, const char *base, const char *rel, char **found_here, int *success);
struct JsonnetJsonValue *CallNative_cgo(void *ctx, const struct JsonnetJsonValue *const *argv, int *success);

#cgo CXXFLAGS: -std=c++0x -O3
*/
import "C"

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"unsafe"
)

type ImportCallback func(base, rel string) (result string, path string, err error)

type NativeCallback func(args ...*JsonValue) (result *JsonValue, err error)

type nativeFunc struct {
	vm       *VM
	argc     int
	callback NativeCallback
}

// Global registry of native functions.  Cgo pointer rules don't allow
// us to pass go pointers directly (may not be stable), so pass uintptr
// keys into this indirect map instead.
var nativeFuncsMu sync.Mutex
var nativeFuncsIdx uintptr
var nativeFuncs = make(map[uintptr]*nativeFunc)

func registerFunc(vm *VM, arity int, callback NativeCallback) uintptr {
	f := nativeFunc{vm: vm, argc: arity, callback: callback}

	nativeFuncsMu.Lock()
	defer nativeFuncsMu.Unlock()

	nativeFuncsIdx++
	for nativeFuncs[nativeFuncsIdx] != nil {
		nativeFuncsIdx++
	}

	nativeFuncs[nativeFuncsIdx] = &f
	return nativeFuncsIdx
}

func getFunc(key uintptr) *nativeFunc {
	nativeFuncsMu.Lock()
	defer nativeFuncsMu.Unlock()

	return nativeFuncs[key]
}

func unregisterFuncs(vm *VM) {
	nativeFuncsMu.Lock()
	defer nativeFuncsMu.Unlock()

	// This is inefficient if there are many
	// simultaneously-existing VMs...
	for idx, f := range nativeFuncs {
		if f.vm == vm {
			delete(nativeFuncs, idx)
		}
	}
}

type VM struct {
	guts           *C.struct_JsonnetVm
	importCallback ImportCallback
}

//export go_call_native
func go_call_native(key uintptr, argv **C.struct_JsonnetJsonValue, okPtr *C.int) *C.struct_JsonnetJsonValue {
	f := getFunc(key)
	vm := f.vm

	goArgv := make([]*JsonValue, f.argc)
	for i := 0; i < f.argc; i++ {
		p := unsafe.Pointer(uintptr(unsafe.Pointer(argv)) + unsafe.Sizeof(*argv)*uintptr(i))
		argptr := (**C.struct_JsonnetJsonValue)(p)
		// NB: argv will be freed by jsonnet after this
		// function exits, so don't want (*JsonValue).destroy
		// finalizer.
		goArgv[i] = &JsonValue{
			vm:   vm,
			guts: *argptr,
		}
	}

	ret, err := f.callback(goArgv...)
	if err != nil {
		*okPtr = C.int(0)
		ret = vm.NewString(err.Error())
	} else {
		*okPtr = C.int(1)
	}

	return ret.take()
}

//export go_call_import
func go_call_import(vmPtr unsafe.Pointer, base, rel *C.char, pathPtr **C.char, okPtr *C.int) *C.char {
	vm := (*VM)(vmPtr)
	result, path, err := vm.importCallback(C.GoString(base), C.GoString(rel))
	if err != nil {
		*okPtr = C.int(0)
		return jsonnetString(vm, err.Error())
	}
	*pathPtr = jsonnetString(vm, path)
	*okPtr = C.int(1)
	return jsonnetString(vm, result)
}

// Evaluate a file containing Jsonnet code, return a JSON string.
func Version() string {
	return C.GoString(C.jsonnet_version())
}

// Create a new Jsonnet virtual machine.
func Make() *VM {
	vm := &VM{guts: C.jsonnet_make()}
	return vm
}

// Complement of Make().
func (vm *VM) Destroy() {
	unregisterFuncs(vm)
	C.jsonnet_destroy(vm.guts)
	vm.guts = nil
}

// jsonnet often wants char* strings that were allocated via
// jsonnet_realloc.  This function does that.
func jsonnetString(vm *VM, s string) *C.char {
	clen := C.size_t(len(s)) + 1 // num bytes including trailing \0

	// TODO: remove additional copy
	cstr := C.CString(s)
	defer C.free(unsafe.Pointer(cstr))

	ret := C.jsonnet_realloc(vm.guts, nil, clen)
	C.memcpy(unsafe.Pointer(ret), unsafe.Pointer(cstr), clen)

	return ret
}

// Evaluate a file containing Jsonnet code, return a JSON string.
func (vm *VM) EvaluateFile(filename string) (string, error) {
	cfilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cfilename))

	var e C.int
	z := C.GoString(C.jsonnet_evaluate_file(vm.guts, cfilename, &e))
	if e != 0 {
		return "", errors.New(z)
	}
	return z, nil
}

// Evaluate a string containing Jsonnet code, return a JSON string.
func (vm *VM) EvaluateSnippet(filename, snippet string) (string, error) {
	cfilename := C.CString(filename)
	csnippet := C.CString(snippet)
	defer func() {
		C.free(unsafe.Pointer(csnippet))
		C.free(unsafe.Pointer(cfilename))
	}()

	var e C.int
	z := C.GoString(C.jsonnet_evaluate_snippet(vm.guts, cfilename, csnippet, &e))
	if e != 0 {
		return "", errors.New(z)
	}
	return z, nil
}

// Format a file containing Jsonnet code, return a JSON string.
func (vm *VM) FormatFile(filename string) (string, error) {
	cfilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cfilename))

	var e C.int
	z := C.GoString(C.jsonnet_fmt_file(vm.guts, cfilename, &e))
	if e != 0 {
		return "", errors.New(z)
	}
	return z, nil
}

// Indentation level when reformatting (number of spaces)
func (vm *VM) FormatIndent(n int) {
	C.jsonnet_fmt_indent(vm.guts, C.int(n))
}

// Format a string containing Jsonnet code, return a JSON string.
func (vm *VM) FormatSnippet(filename, snippet string) (string, error) {
	cfilename := C.CString(filename)
	csnippet := C.CString(snippet)
	defer func() {
		C.free(unsafe.Pointer(csnippet))
		C.free(unsafe.Pointer(cfilename))
	}()

	var e C.int
	z := C.GoString(C.jsonnet_fmt_snippet(vm.guts, cfilename, csnippet, &e))
	if e != 0 {
		return "", errors.New(z)
	}
	return z, nil
}

// Override the callback used to locate imports.
func (vm *VM) ImportCallback(f ImportCallback) {
	vm.importCallback = f
	C.jsonnet_import_callback(
		vm.guts,
		(*C.JsonnetImportCallback)(unsafe.Pointer(C.CallImport_cgo)),
		unsafe.Pointer(vm))
}

// NativeCallback is a helper around NativeCallbackRaw that uses
// `reflect` to convert argument and result types to/from JsonValue.
// `f` is expected to be a function that takes argument types
// supported by `(*JsonValue).Extract` and returns `(x, error)` where
// `x` is a type supported by `NewJson`.
func (vm *VM) NativeCallback(name string, params []string, f interface{}) {
	ty := reflect.TypeOf(f)
	if ty.NumIn() != len(params) {
		panic("Wrong number of parameters")
	}
	if ty.NumOut() != 2 {
		panic("Wrong number of output parameters")
	}

	wrapper := func(args ...*JsonValue) (*JsonValue, error) {
		in := make([]reflect.Value, len(args))
		for i, arg := range args {
			value := reflect.ValueOf(arg.Extract())
			if vty := value.Type(); !vty.ConvertibleTo(ty.In(i)) {
				return nil, fmt.Errorf("parameter %d (type %s) cannot be converted to type %s", i, vty, ty.In(i))
			}
			in[i] = value.Convert(ty.In(i))
		}

		out := reflect.ValueOf(f).Call(in)

		result := vm.NewJson(out[0].Interface())
		var err error
		if out[1].IsValid() && !out[1].IsNil() {
			err = out[1].Interface().(error)
		}
		return result, err
	}

	vm.NativeCallbackRaw(name, params, wrapper)
}

func (vm *VM) NativeCallbackRaw(name string, params []string, f NativeCallback) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	// jsonnet expects this to be NULL-terminated, so the last
	// element is left as nil
	cparams := make([]*C.char, len(params)+1)
	for i, param := range params {
		cparams[i] = C.CString(param)
		defer C.free(unsafe.Pointer(cparams[i]))
	}

	key := registerFunc(vm, len(params), f)
	C.jsonnet_native_callback(
		vm.guts,
		cname,
		(*C.JsonnetNativeCallback)(C.CallNative_cgo),
		unsafe.Pointer(key),
		(**C.char)(unsafe.Pointer(&cparams[0])))
}

// Bind a Jsonnet external var to the given value.
func (vm *VM) ExtVar(key, val string) {
	ckey := C.CString(key)
	cval := C.CString(val)
	defer func() {
		C.free(unsafe.Pointer(cval))
		C.free(unsafe.Pointer(ckey))
	}()
	C.jsonnet_ext_var(vm.guts, ckey, cval)
}

// Bind a Jsonnet external var to the given Jsonnet code.
func (vm *VM) ExtCode(key, val string) {
	ckey := C.CString(key)
	cval := C.CString(val)
	defer func() {
		C.free(unsafe.Pointer(cval))
		C.free(unsafe.Pointer(ckey))
	}()
	C.jsonnet_ext_code(vm.guts, ckey, cval)
}

// Bind a Jsonnet top-level argument to the given value.
func (vm *VM) TlaVar(key, val string) {
	ckey := C.CString(key)
	cval := C.CString(val)
	defer func() {
		C.free(unsafe.Pointer(cval))
		C.free(unsafe.Pointer(ckey))
	}()
	C.jsonnet_tla_var(vm.guts, ckey, cval)
}

// Bind a Jsonnet top-level argument to the given Jsonnet code.
func (vm *VM) TlaCode(key, val string) {
	ckey := C.CString(key)
	cval := C.CString(val)
	defer func() {
		C.free(unsafe.Pointer(cval))
		C.free(unsafe.Pointer(ckey))
	}()
	C.jsonnet_tla_code(vm.guts, ckey, cval)
}

// Set the maximum stack depth.
func (vm *VM) MaxStack(v uint) {
	C.jsonnet_max_stack(vm.guts, C.uint(v))
}

// Set the number of lines of stack trace to display (0 for all of them).
func (vm *VM) MaxTrace(v uint) {
	C.jsonnet_max_trace(vm.guts, C.uint(v))
}

// Set the number of objects required before a garbage collection cycle is allowed.
func (vm *VM) GcMinObjects(v uint) {
	C.jsonnet_gc_min_objects(vm.guts, C.uint(v))
}

// Run the garbage collector after this amount of growth in the number of objects.
func (vm *VM) GcGrowthTrigger(v float64) {
	C.jsonnet_gc_growth_trigger(vm.guts, C.double(v))
}

// Expect a string as output and don't JSON encode it.
func (vm *VM) StringOutput(v bool) {
	if v {
		C.jsonnet_string_output(vm.guts, C.int(1))
	} else {
		C.jsonnet_string_output(vm.guts, C.int(0))
	}
}

// Add to the default import callback's library search path.
func (vm *VM) JpathAdd(path string) {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))
	C.jsonnet_jpath_add(vm.guts, cpath)
}

/* The following are not implemented because they are trivial to implement in Go on top of the
 * existing API by parsing and post-processing the JSON output by regular evaluation.
 *
 * jsonnet_evaluate_file_multi
 * jsonnet_evaluate_snippet_multi
 * jsonnet_evaluate_file_stream
 * jsonnet_evaluate_snippet_stream
 */

// JsonValue represents a jsonnet JSON object.
type JsonValue struct {
	vm   *VM
	guts *C.struct_JsonnetJsonValue
}

func (v *JsonValue) Extract() interface{} {
	if x, ok := v.ExtractString(); ok {
		return x
	}
	if x, ok := v.ExtractNumber(); ok {
		return x
	}
	if x, ok := v.ExtractBool(); ok {
		return x
	}
	if ok := v.ExtractNull(); ok {
		return nil
	}
	panic("Unable to extract value")
}

// ExtractString returns the string value and true if the value was a string
func (v *JsonValue) ExtractString() (string, bool) {
	cstr := C.jsonnet_json_extract_string(v.vm.guts, v.guts)
	if cstr == nil {
		return "", false
	}
	return C.GoString(cstr), true
}

func (v *JsonValue) ExtractNumber() (float64, bool) {
	var ret C.double
	ok := C.jsonnet_json_extract_number(v.vm.guts, v.guts, &ret)
	return float64(ret), ok != 0
}

func (v *JsonValue) ExtractBool() (bool, bool) {
	ret := C.jsonnet_json_extract_bool(v.vm.guts, v.guts)
	switch ret {
	case 0:
		return false, true
	case 1:
		return true, true
	case 2:
		// Not a bool
		return false, false
	default:
		panic("jsonnet_json_extract_number returned unexpected value")
	}
}

// ExtractNull returns true iff the value is null
func (v *JsonValue) ExtractNull() bool {
	ret := C.jsonnet_json_extract_null(v.vm.guts, v.guts)
	return ret != 0
}

func (vm *VM) newjson(ptr *C.struct_JsonnetJsonValue) *JsonValue {
	v := &JsonValue{vm: vm, guts: ptr}
	runtime.SetFinalizer(v, (*JsonValue).destroy)
	return v
}

func (v *JsonValue) destroy() {
	if v.guts == nil {
		return
	}
	C.jsonnet_json_destroy(v.vm.guts, v.guts)
	v.guts = nil
	runtime.SetFinalizer(v, nil)
}

// Take ownership of the embedded ptr, effectively consuming the JsonValue
func (v *JsonValue) take() *C.struct_JsonnetJsonValue {
	ptr := v.guts
	if ptr == nil {
		panic("taking nil pointer from JsonValue")
	}
	v.guts = nil
	runtime.SetFinalizer(v, nil)
	return ptr
}

func (vm *VM) NewJson(value interface{}) *JsonValue {
	switch val := value.(type) {
	case string:
		return vm.NewString(val)
	case int:
		return vm.NewNumber(float64(val))
	case float64:
		return vm.NewNumber(val)
	case bool:
		return vm.NewBool(val)
	case nil:
		return vm.NewNull()
	case []interface{}:
		a := vm.NewArray()
		for _, v := range val {
			a.ArrayAppend(vm.NewJson(v))
		}
		return a
	case map[string]interface{}:
		o := vm.NewObject()
		for k, v := range val {
			o.ObjectAppend(k, vm.NewJson(v))
		}
		return o
	default:
		panic(fmt.Sprintf("NewJson can't handle type: %T", value))
	}
}

func (vm *VM) NewString(v string) *JsonValue {
	cstr := C.CString(v)
	defer C.free(unsafe.Pointer(cstr))
	ptr := C.jsonnet_json_make_string(vm.guts, cstr)
	return vm.newjson(ptr)
}

func (vm *VM) NewNumber(v float64) *JsonValue {
	ptr := C.jsonnet_json_make_number(vm.guts, C.double(v))
	return vm.newjson(ptr)
}

func (vm *VM) NewBool(v bool) *JsonValue {
	var i C.int
	if v {
		i = 1
	} else {
		i = 0
	}
	ptr := C.jsonnet_json_make_bool(vm.guts, i)
	return vm.newjson(ptr)
}

func (vm *VM) NewNull() *JsonValue {
	ptr := C.jsonnet_json_make_null(vm.guts)
	return vm.newjson(ptr)
}

func (vm *VM) NewArray() *JsonValue {
	ptr := C.jsonnet_json_make_array(vm.guts)
	return vm.newjson(ptr)
}

func (v *JsonValue) ArrayAppend(item *JsonValue) {
	C.jsonnet_json_array_append(v.vm.guts, v.guts, item.take())
}

func (vm *VM) NewObject() *JsonValue {
	ptr := C.jsonnet_json_make_object(vm.guts)
	return vm.newjson(ptr)
}

func (v *JsonValue) ObjectAppend(key string, value *JsonValue) {
	ckey := C.CString(key)
	defer C.free(unsafe.Pointer(ckey))

	C.jsonnet_json_object_append(v.vm.guts, v.guts, ckey, value.take())
}
