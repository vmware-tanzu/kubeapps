// Copyright 2017 The kubecfg authors
//
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package cmd

import (
	"bytes"
	"encoding/json"
	goflag "flag"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/ksonnet/kubecfg/metadata"
	"github.com/ksonnet/kubecfg/template"
	"github.com/ksonnet/kubecfg/utils"

	// Register auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

const (
	flagVerbose    = "verbose"
	flagJpath      = "jpath"
	flagExtVar     = "ext-str"
	flagExtVarFile = "ext-str-file"
	flagTlaVar     = "tla-str"
	flagTlaVarFile = "tla-str-file"
	flagResolver   = "resolve-images"
	flagResolvFail = "resolve-images-error"
	flagAPISpec    = "api-spec"

	// For use in the commands (e.g., diff, apply, delete) that require either an
	// environment or the -f flag.
	flagFile      = "file"
	flagFileShort = "f"

	componentsExtCodeKey = "__ksonnet/components"
)

var clientConfig clientcmd.ClientConfig
var overrides clientcmd.ConfigOverrides
var loadingRules clientcmd.ClientConfigLoadingRules

func init() {
	RootCmd.PersistentFlags().CountP(flagVerbose, "v", "Increase verbosity. May be given multiple times.")

	// The "usual" clientcmd/kubectl flags
	loadingRules = *clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	clientConfig = clientcmd.NewInteractiveDeferredLoadingClientConfig(&loadingRules, &overrides, os.Stdin)

	RootCmd.PersistentFlags().Set("logtostderr", "true")
}

func bindJsonnetFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringSliceP(flagJpath, "J", nil, "Additional jsonnet library search path")
	cmd.PersistentFlags().StringSliceP(flagExtVar, "V", nil, "Values of external variables")
	cmd.PersistentFlags().StringSlice(flagExtVarFile, nil, "Read external variable from a file")
	cmd.PersistentFlags().StringSliceP(flagTlaVar, "A", nil, "Values of top level arguments")
	cmd.PersistentFlags().StringSlice(flagTlaVarFile, nil, "Read top level argument from a file")
	cmd.PersistentFlags().String(flagResolver, "noop", "Change implementation of resolveImage native function. One of: noop, registry")
	cmd.PersistentFlags().String(flagResolvFail, "warn", "Action when resolveImage fails. One of ignore,warn,error")
}

func bindClientGoFlags(cmd *cobra.Command) {
	kflags := clientcmd.RecommendedConfigOverrideFlags("")
	ep := &loadingRules.ExplicitPath
	cmd.PersistentFlags().StringVar(ep, "kubeconfig", "", "Path to a kube config. Only required if out-of-cluster")
	clientcmd.BindOverrideFlags(&overrides, cmd.PersistentFlags(), kflags)
}

// RootCmd is the root of cobra subcommand tree
var RootCmd = &cobra.Command{
	Use:           "kubecfg",
	Short:         "Synchronise Kubernetes resources with config files",
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		goflag.CommandLine.Parse([]string{})
		flags := cmd.Flags()
		out := cmd.OutOrStderr()
		log.SetOutput(out)

		logFmt := NewLogFormatter(out)
		log.SetFormatter(logFmt)

		verbosity, err := flags.GetCount(flagVerbose)
		if err != nil {
			return err
		}
		log.SetLevel(logLevel(verbosity))

		return nil
	},
}

// clientConfig.Namespace() is broken in client-go 3.0:
// namespace in config erroneously overrides explicit --namespace
func defaultNamespace() (string, error) {
	if overrides.Context.Namespace != "" {
		return overrides.Context.Namespace, nil
	}
	ns, _, err := clientConfig.Namespace()
	return ns, err
}

func logLevel(verbosity int) log.Level {
	switch verbosity {
	case 0:
		return log.InfoLevel
	default:
		return log.DebugLevel
	}
}

type logFormatter struct {
	escapes  *terminal.EscapeCodes
	colorise bool
}

// NewLogFormatter creates a new log.Formatter customised for writer
func NewLogFormatter(out io.Writer) log.Formatter {
	var ret = logFormatter{}
	if f, ok := out.(*os.File); ok {
		ret.colorise = terminal.IsTerminal(int(f.Fd()))
		ret.escapes = terminal.NewTerminal(f, "").Escape
	}
	return &ret
}

func (f *logFormatter) levelEsc(level log.Level) []byte {
	switch level {
	case log.DebugLevel:
		return []byte{}
	case log.WarnLevel:
		return f.escapes.Yellow
	case log.ErrorLevel, log.FatalLevel, log.PanicLevel:
		return f.escapes.Red
	default:
		return f.escapes.Blue
	}
}

func (f *logFormatter) Format(e *log.Entry) ([]byte, error) {
	buf := bytes.Buffer{}
	if f.colorise {
		buf.Write(f.levelEsc(e.Level))
		fmt.Fprintf(&buf, "%-5s ", strings.ToUpper(e.Level.String()))
		buf.Write(f.escapes.Reset)
	}

	buf.WriteString(strings.TrimSpace(e.Message))
	buf.WriteString("\n")

	return buf.Bytes(), nil
}

func newExpander(cmd *cobra.Command) (*template.Expander, error) {
	flags := cmd.Flags()
	spec := template.Expander{}
	var err error

	spec.EnvJPath = filepath.SplitList(os.Getenv("KUBECFG_JPATH"))

	spec.FlagJpath, err = flags.GetStringSlice(flagJpath)
	if err != nil {
		return nil, err
	}

	spec.ExtVars, err = flags.GetStringSlice(flagExtVar)
	if err != nil {
		return nil, err
	}

	spec.ExtVarFiles, err = flags.GetStringSlice(flagExtVarFile)
	if err != nil {
		return nil, err
	}

	spec.TlaVars, err = flags.GetStringSlice(flagTlaVar)
	if err != nil {
		return nil, err
	}

	spec.TlaVarFiles, err = flags.GetStringSlice(flagTlaVarFile)
	if err != nil {
		return nil, err
	}

	spec.Resolver, err = flags.GetString(flagResolver)
	if err != nil {
		return nil, err
	}
	spec.FailAction, err = flags.GetString(flagResolvFail)
	if err != nil {
		return nil, err
	}

	return &spec, nil
}

// For debugging
func dumpJSON(v interface{}) string {
	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return err.Error()
	}
	return string(buf.Bytes())
}

func restClientPool(cmd *cobra.Command, envName *string) (dynamic.ClientPool, discovery.DiscoveryInterface, error) {
	if envName != nil {
		err := overrideCluster(*envName)
		if err != nil {
			return nil, nil, err
		}
	}

	conf, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, nil, err
	}

	disco, err := discovery.NewDiscoveryClientForConfig(conf)
	if err != nil {
		return nil, nil, err
	}

	discoCache := utils.NewMemcachedDiscoveryClient(disco)
	mapper := discovery.NewDeferredDiscoveryRESTMapper(discoCache, dynamic.VersionInterfaces)
	pathresolver := dynamic.LegacyAPIPathResolverFunc

	pool := dynamic.NewClientPool(conf, mapper, pathresolver)
	return pool, discoCache, nil
}

type envSpec struct {
	env   *string
	files []string
}

// addEnvCmdFlags adds the flags that are common to the family of commands
// whose form is `[<env>|-f <file-name>]`, e.g., `apply` and `delete`.
func addEnvCmdFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringArrayP(flagFile, flagFileShort, nil, "Filename or directory that contains the configuration to apply (accepts YAML, JSON, and Jsonnet)")
}

// parseEnvCmd parses the family of commands that come in the form `[<env>|-f
// <file-name>]`, e.g., `apply` and `delete`.
func parseEnvCmd(cmd *cobra.Command, args []string) (*envSpec, error) {
	flags := cmd.Flags()

	files, err := flags.GetStringArray(flagFile)
	if err != nil {
		return nil, err
	}

	var env *string
	if len(args) == 1 {
		env = &args[0]
	}

	return &envSpec{env: env, files: files}, nil
}

// overrideCluster ensures that the cluster URI specified in the environment is
// associated in the user's kubeconfig file during deployment to a ksonnet
// environment. We will error out if it is not.
//
// If the environment URI the user is attempting to deploy to is not the current
// kubeconfig context, we must manually override the client-go --cluster flag
// to ensure we are deploying to the correct cluster.
func overrideCluster(envName string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	wd := metadata.AbsPath(cwd)

	metadataManager, err := metadata.Find(wd)
	if err != nil {
		return err
	}

	rawConfig, err := clientConfig.RawConfig()
	if err != nil {
		return err
	}

	var clusterURIs = make(map[string]string)
	for name, cluster := range rawConfig.Clusters {
		clusterURIs[cluster.Server] = name
	}

	//
	// check to ensure that the environment we are trying to deploy to is
	// created, and that the environment URI is located in kubeconfig.
	//

	log.Debugf("Validating deployment at '%s' with cluster URIs '%v'", envName, reflect.ValueOf(clusterURIs).MapKeys())
	env, err := metadataManager.GetEnvironment(envName)
	if err != nil {
		return err
	}

	if _, ok := clusterURIs[env.URI]; ok {
		clusterName := clusterURIs[env.URI]
		log.Debugf("Overwriting --cluster flag with '%s'", clusterName)
		overrides.Context.Cluster = clusterName
		return nil
	}

	return fmt.Errorf("Attempting to deploy to environment '%s' at %s, but there are no clusters with that URI", envName, env.URI)
}

// expandEnvCmdObjs finds and expands templates for the family of commands of
// the form `[<env>|-f <file-name>]`, e.g., `apply` and `delete`. That is, if
// the user passes a list of files, we will expand all templates in those files,
// while if a user passes an environment name, we will expand all component
// files using that environment.
func expandEnvCmdObjs(cmd *cobra.Command, envSpec *envSpec, cwd metadata.AbsPath) ([]*unstructured.Unstructured, error) {
	expander, err := newExpander(cmd)
	if err != nil {
		return nil, err
	}

	//
	// Get all filenames that contain templates to expand. Importantly, we need to
	// enforce the form `[<env-name>|-f <file-name>]`; that is, we need to make
	// sure that the user either passed an environment name or a `-f` flag.
	//

	envPresent := envSpec.env != nil
	filesPresent := len(envSpec.files) > 0

	if !envPresent && !filesPresent {
		return nil, fmt.Errorf("Must specify either an environment or a file list, or both")
	}

	fileNames := envSpec.files
	if envPresent {
		manager, err := metadata.Find(cwd)
		if err != nil {
			return nil, err
		}

		libPath, envLibPath := manager.LibPaths(*envSpec.env)
		expander.FlagJpath = append([]string{string(libPath), string(envLibPath)}, expander.FlagJpath...)

		if !filesPresent {
			fileNames, err = manager.ComponentPaths()
			if err != nil {
				return nil, err
			}
			baseObjExtCode := fmt.Sprintf("%s=%s", componentsExtCodeKey, constructBaseObj(fileNames))
			expander.ExtCodes = append([]string{baseObjExtCode})
		}
	}

	//
	// Expand templates.
	//

	return expander.Expand(fileNames)
}

// constructBaseObj constructs the base Jsonnet object that represents k-v
// pairs of component name -> component imports. For example,
//
//   {
//      foo: import "components/foo.jsonnet"
//   }
func constructBaseObj(paths []string) string {
	var obj bytes.Buffer
	obj.WriteString("{\n")
	for _, p := range paths {
		ext := path.Ext(p)
		if path.Ext(p) != ".jsonnet" {
			continue
		}

		name := strings.TrimSuffix(path.Base(p), ext)
		fmt.Fprintf(&obj, "  %s: import \"%s\",\n", name, p)
	}
	obj.WriteString("}\n")
	return obj.String()
}
