package template

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/ksonnet/kubecfg/utils"
	log "github.com/sirupsen/logrus"
	jsonnet "github.com/strickyak/jsonnet_cgo"
)

type Expander struct {
	EnvJPath    []string
	FlagJpath   []string
	ExtVars     []string
	ExtVarFiles []string
	TlaVars     []string
	TlaVarFiles []string
	ExtCodes    []string

	Resolver   string
	FailAction string
}

func (spec *Expander) Expand(paths []string) ([]*unstructured.Unstructured, error) {
	vm, err := spec.jsonnetVM()
	if err != nil {
		return nil, err
	}
	defer vm.Destroy()

	res := []*unstructured.Unstructured{}
	for _, path := range paths {
		objs, err := utils.Read(vm, path)
		if err != nil {
			return nil, fmt.Errorf("Error reading %s: %v", path, err)
		}
		res = append(res, utils.FlattenToV1(objs)...)
	}
	return res, nil
}

// JsonnetVM constructs a new jsonnet.VM, according to command line
// flags
func (spec *Expander) jsonnetVM() (*jsonnet.VM, error) {
	vm := jsonnet.Make()

	for _, p := range spec.EnvJPath {
		log.Debugln("Adding jsonnet search path", p)
		vm.JpathAdd(p)
	}

	for _, p := range spec.FlagJpath {
		log.Debugln("Adding jsonnet search path", p)
		vm.JpathAdd(p)
	}

	for _, extvar := range spec.ExtVars {
		kv := strings.SplitN(extvar, "=", 2)
		switch len(kv) {
		case 1:
			v, present := os.LookupEnv(kv[0])
			if present {
				vm.ExtVar(kv[0], v)
			} else {
				return nil, fmt.Errorf("Missing environment variable: %s", kv[0])
			}
		case 2:
			vm.ExtVar(kv[0], kv[1])
		}
	}

	for _, extvar := range spec.ExtVarFiles {
		kv := strings.SplitN(extvar, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("Failed to parse ext var files: missing '=' in %s", extvar)
		}
		v, err := ioutil.ReadFile(kv[1])
		if err != nil {
			return nil, err
		}
		vm.ExtVar(kv[0], string(v))
	}

	for _, tlavar := range spec.TlaVars {
		kv := strings.SplitN(tlavar, "=", 2)
		switch len(kv) {
		case 1:
			v, present := os.LookupEnv(kv[0])
			if present {
				vm.TlaVar(kv[0], v)
			} else {
				return nil, fmt.Errorf("Missing environment variable: %s", kv[0])
			}
		case 2:
			vm.TlaVar(kv[0], kv[1])
		}
	}

	for _, tlavar := range spec.TlaVarFiles {
		kv := strings.SplitN(tlavar, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("Failed to parse tla var files: missing '=' in %s", tlavar)
		}
		v, err := ioutil.ReadFile(kv[1])
		if err != nil {
			return nil, err
		}
		vm.TlaVar(kv[0], string(v))
	}

	for _, extcode := range spec.ExtCodes {
		kv := strings.SplitN(extcode, "=", 2)
		switch len(kv) {
		case 1:
			v, present := os.LookupEnv(kv[0])
			if present {
				vm.ExtCode(kv[0], v)
			} else {
				return nil, fmt.Errorf("Missing environment variable: %s", kv[0])
			}
		case 2:
			vm.ExtCode(kv[0], kv[1])
		}
	}

	resolver, err := spec.buildResolver()
	if err != nil {
		return nil, err
	}
	utils.RegisterNativeFuncs(vm, resolver)

	return vm, nil
}
