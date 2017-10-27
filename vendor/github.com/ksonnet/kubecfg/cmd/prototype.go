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
	"fmt"
	"os"
	"strings"

	"github.com/spf13/pflag"

	"github.com/ksonnet/kubecfg/metadata"
	"github.com/ksonnet/kubecfg/prototype"
	"github.com/ksonnet/kubecfg/prototype/snippet"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(prototypeCmd)
	prototypeCmd.AddCommand(prototypeListCmd)
	prototypeCmd.AddCommand(prototypeDescribeCmd)
	prototypeCmd.AddCommand(prototypeSearchCmd)
	prototypeCmd.AddCommand(prototypeUseCmd)
	prototypeCmd.AddCommand(prototypePreviewCmd)
}

var prototypeCmd = &cobra.Command{
	Use:   "prototype",
	Short: `Instantiate, inspect, and get examples for ksonnet prototypes`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("Command 'prototype' requires a subcommand\n\n%s", cmd.UsageString())
	},
	Long: `Manage, inspect, instantiate, and get examples for ksonnet prototypes.

Prototypes are Kubernetes app configuration templates with "holes" that can be
filled in by (e.g.) the ksonnet CLI tool or a language server. For example, a
prototype for a 'apps.v1beta1.Deployment' might require a name and image, and
the ksonnet CLI could expand this to a fully-formed 'Deployment' object.

Commands:
  use      Instantiate prototype, filling in parameters from flags, and
           emitting the generated code to stdout.
  describe Display documentation and details about a prototype
  search   Search for a prototype`,

	Example: `  # Display documentation about prototype
  # 'io.ksonnet.pkg.prototype.simple-deployment', including:
  #
  #   (1) a description of what gets generated during instantiation
  #   (2) a list of parameters that are required to be passed in with CLI flags
  #
  # NOTE: Many subcommands only require the user to specify enough of the
  # identifier to disambiguate it among other known prototypes, which is why
  # 'simple-deployment' is given as argument instead of the fully-qualified
  # name.
  ksonnet prototype describe simple-deployment

  # Instantiate prototype 'io.ksonnet.pkg.prototype.simple-deployment', using
  # the 'nginx' image, and port 80 exposed.
  #
  # SEE ALSO: Note above for a description of why this subcommand can take
  # 'simple-deployment' instead of the fully-qualified prototype name.
  ksonnet prototype use simple-deployment \
    --name=nginx                          \
    --image=nginx                         \
    --port=80                             \
    --portName=http

  # Search known prototype metadata for the string 'deployment'.
  ksonnet prototype search deployment`,
}

var prototypeListCmd = &cobra.Command{
	Use:   "list <name-substring>",
	Short: `List all known ksonnet prototypes`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			return fmt.Errorf("Command 'prototype list' does not take any arguments")
		}

		index := prototype.NewIndex([]*prototype.SpecificationSchema{})
		protos, err := index.List()
		if err != nil {
			return err
		} else if len(protos) == 0 {
			return fmt.Errorf("No prototypes found")
		}

		fmt.Print(protos)

		return nil
	},
	Long: `List all known ksonnet prototypes.`,
}

var prototypeDescribeCmd = &cobra.Command{
	Use:   "describe <prototype-name>",
	Short: `Describe a ksonnet prototype`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("Command 'prototype describe' requires a prototype name\n\n%s", cmd.UsageString())
		}

		query := args[0]

		proto, err := fundUniquePrototype(query)
		if err != nil {
			return err
		}

		fmt.Println(`PROTOTYPE NAME:`)
		fmt.Println(proto.Name)
		fmt.Println()
		fmt.Println(`DESCRIPTION:`)
		fmt.Println(proto.Template.Description)
		fmt.Println()
		fmt.Println(`REQUIRED PARAMETERS:`)
		fmt.Println(proto.RequiredParams().PrettyString("  "))
		fmt.Println()
		fmt.Println(`OPTIONAL PARAMETERS:`)
		fmt.Println(proto.OptionalParams().PrettyString("  "))
		fmt.Println()
		fmt.Println(`TEMPLATE TYPES AVAILABLE:`)
		fmt.Println(fmt.Sprintf("  %s", proto.Template.AvailableTemplates()))

		return nil
	},
	Long: `Output documentation, examples, and other information for some ksonnet
prototype uniquely identified by some (possibly partial) 'prototype-name'. This
includes:

  (1) a description of what gets generated during instantiation
  (2) a list of parameters that are required to be passed in with CLI flags

'prototype-name' need only contain enough of the suffix of a name to uniquely
disambiguate it among known names. For example, 'deployment' may resolve
ambiguously, in which case 'use' will fail, while 'simple-deployment' might be
unique enough to resolve to 'io.ksonnet.pkg.prototype.simple-deployment'.`,

	Example: `  # Display documentation about prototype, including:
  ksonnet prototype describe io.ksonnet.pkg.prototype.simple-deployment

  # Display documentation about prototype using a unique suffix of an
  # identifier. That is, this command only requires a long enough suffix to
  # uniquely identify a ksonnet prototype. In this example, the suffix
  # 'simple-deployment' is enough to uniquely identify
  # 'io.ksonnet.pkg.prototype.simple-deployment', but 'deployment' might not
  # be, as several names end with that suffix.
  ksonnet prototype describe simple-deployment`,
}

var prototypeSearchCmd = &cobra.Command{
	Use:   "search <name-substring>",
	Short: `Search for a ksonnet prototype`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("Command 'prototype search' requires a prototype name\n\n%s", cmd.UsageString())
		}

		query := args[0]

		index := prototype.NewIndex([]*prototype.SpecificationSchema{})
		protos, err := index.SearchNames(query, prototype.Substring)
		if err != nil {
			return err
		} else if len(protos) == 0 {
			return fmt.Errorf("Failed to find any search results for query '%s'", query)
		}

		fmt.Print(protos)

		return nil
	},
	Long: `Search ksonnet for prototypes whose names contain 'name-substring'.`,
	Example: `  # Search known prototype metadata for the string 'deployment'.
  ksonnet prototype search deployment`,
}

var prototypePreviewCmd = &cobra.Command{
	Use:                "preview <prototype-name> [type] [parameter-flags]",
	Short:              `Expand prototype, emitting the generated code to stdout`,
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, rawArgs []string) error {
		if len(rawArgs) < 1 {
			return fmt.Errorf("Command 'prototype preview' requires a prototype name\n\n%s", cmd.UsageString())
		}

		query := rawArgs[0]

		proto, err := fundUniquePrototype(query)
		if err != nil {
			return err
		}

		bindPrototypeFlags(cmd, proto)

		cmd.DisableFlagParsing = false
		err = cmd.ParseFlags(rawArgs)
		if err != nil {
			return err
		}
		flags := cmd.Flags()

		// Try to find the template type (if it is supplied) after the args are
		// parsed. Note that the case that `len(args) == 0` is handled at the
		// beginning of this command.
		var templateType prototype.TemplateType
		if args := flags.Args(); len(args) == 1 {
			templateType = prototype.Jsonnet
		} else if len(args) == 2 {
			templateType, err = prototype.ParseTemplateType(args[1])
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("Incorrect number of arguments supplied to 'prototype preview'\n\n%s", cmd.UsageString())
		}

		text, err := expandPrototype(proto, flags, templateType)
		if err != nil {
			return err
		}

		fmt.Println(text)
		return nil
	},
	Long: `Expand prototype uniquely identified by (possibly partial)
'prototype-name', filling in parameters from flags, and emitting the generated
code to stdout.

Note also that 'prototype-name' need only contain enough of the suffix of a name
to uniquely disambiguate it among known names. For example, 'deployment' may
resolve ambiguously, in which case 'use' will fail, while 'deployment' might be
unique enough to resolve to 'io.ksonnet.pkg.single-port-deployment'.`,

	Example: `  # Preview prototype 'io.ksonnet.pkg.single-port-deployment', using the
  # 'nginx' image, and port 80 exposed.
  ksonnet prototype preview io.ksonnet.pkg.prototype.simple-deployment \
    --name=nginx                                                       \
    --image=nginx

  # Preview prototype using a unique suffix of an identifier. See
  # introduction of help message for more information on how this works.
  ksonnet prototype preview simple-deployment \
    --name=nginx                              \
    --image=nginx

  # Preview prototype 'io.ksonnet.pkg.single-port-deployment' as YAML,
  # placing the result in 'components/nginx-depl.yaml. Note that some templates
  # do not have a YAML or JSON versions.
  ksonnet prototype preview deployment nginx-depl yaml \
    --name=nginx                                       \
    --image=nginx`,
}

var prototypeUseCmd = &cobra.Command{
	Use:                "use <prototype-name> <componentName> [type] [parameter-flags]",
	Short:              `Expand prototype, place in components/ directory of ksonnet app`,
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, rawArgs []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		manager, err := metadata.Find(metadata.AbsPath(cwd))
		if err != nil {
			return fmt.Errorf("'prototype use' can only be run in a ksonnet application directory:\n\n%v", err)
		}

		if len(rawArgs) < 1 {
			return fmt.Errorf("Command 'prototype preview' requires a prototype name\n\n%s", cmd.UsageString())
		}

		query := rawArgs[0]

		proto, err := fundUniquePrototype(query)
		if err != nil {
			return err
		}

		bindPrototypeFlags(cmd, proto)

		cmd.DisableFlagParsing = false
		err = cmd.ParseFlags(rawArgs)
		if err != nil {
			return err
		}
		flags := cmd.Flags()

		// Try to find the template type (if it is supplied) after the args are
		// parsed. Note that the case that `len(args) == 0` is handled at the
		// beginning of this command.
		var componentName string
		var templateType prototype.TemplateType
		if args := flags.Args(); len(args) == 1 {
			return fmt.Errorf("'prototype use' is missing argument 'componentName'\n\n%s", cmd.UsageString())
		} else if len(args) == 2 {
			componentName = args[1]
			templateType = prototype.Jsonnet
		} else if len(args) == 3 {
			componentName = args[1]
			templateType, err = prototype.ParseTemplateType(args[1])
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("'prototype use' has too many arguments (takes a prototype name and a component name)\n\n%s", cmd.UsageString())
		}

		text, err := expandPrototype(proto, flags, templateType)
		if err != nil {
			return err
		}

		return manager.CreateComponent(componentName, text, templateType)
	},
	Long: `Expand prototype uniquely identified by (possibly partial) 'prototype-name',
filling in parameters from flags, and placing it into the file
'components/componentName', with the appropriate extension set. For example, the
following command will expand template 'io.ksonnet.pkg.single-port-deployment'
and place it in the file 'components/nginx-depl.jsonnet' (since by default
ksonnet will expand templates as Jsonnet).

  ksonnet prototype use io.ksonnet.pkg.single-port-deployment nginx-depl \
    --name=nginx                                                         \
    --image=nginx

Note that if we were to specify to expand the template as JSON or YAML, we would
generate a file with a '.json' or '.yaml' extension, respectively. See examples
below for an example of how to do this.

Note also that 'prototype-name' need only contain enough of the suffix of a name
to uniquely disambiguate it among known names. For example, 'deployment' may
resolve ambiguously, in which case 'use' will fail, while 'deployment' might be
unique enough to resolve to 'io.ksonnet.pkg.single-port-deployment'.`,

	Example: `  # Instantiate prototype 'io.ksonnet.pkg.single-port-deployment', using the
  # 'nginx' image. The expanded prototype is placed in
  # 'components/nginx-depl.jsonnet'.
  ksonnet prototype use io.ksonnet.pkg.prototype.simple-deployment nginx-depl \
    --name=nginx                                                              \
    --image=nginx

  # Instantiate prototype 'io.ksonnet.pkg.single-port-deployment' using the
  # unique suffix, 'deployment'. The expanded prototype is again placed in
  # 'components/nginx-depl.jsonnet'. See introduction of help message for more
  # information on how this works. Note that if you have imported another
  # prototype with this suffix, this may resolve ambiguously for you.
  ksonnet prototype use deployment nginx-depl \
    --name=nginx                              \
    --image=nginx

  # Instantiate prototype 'io.ksonnet.pkg.single-port-deployment' as YAML,
  # placing the result in 'components/nginx-depl.yaml. Note that some templates
  # do not have a YAML or JSON versions.
  ksonnet prototype use deployment nginx-depl yaml \
    --name=nginx                              \
    --image=nginx`,
}

func bindPrototypeFlags(cmd *cobra.Command, proto *prototype.SpecificationSchema) {
	for _, param := range proto.RequiredParams() {
		cmd.PersistentFlags().String(param.Name, "", param.Description)
	}

	for _, param := range proto.OptionalParams() {
		cmd.PersistentFlags().String(param.Name, *param.Default, param.Description)
	}
}

func expandPrototype(proto *prototype.SpecificationSchema, flags *pflag.FlagSet, templateType prototype.TemplateType) (string, error) {
	missingReqd := prototype.ParamSchemas{}
	values := map[string]string{}
	for _, param := range proto.RequiredParams() {
		val, err := flags.GetString(param.Name)
		if err != nil {
			return "", err
		} else if val == "" {
			missingReqd = append(missingReqd, param)
		} else if _, ok := values[param.Name]; ok {
			return "", fmt.Errorf("Prototype '%s' has multiple parameters with name '%s'", proto.Name, param.Name)
		}

		quoted, err := param.Quote(val)
		if err != nil {
			return "", err
		}
		values[param.Name] = quoted
	}

	if len(missingReqd) > 0 {
		return "", fmt.Errorf("Failed to instantiate prototype '%s'. The following required parameters are missing:\n%s", proto.Name, missingReqd.PrettyString(""))
	}

	for _, param := range proto.OptionalParams() {
		val, err := flags.GetString(param.Name)
		if err != nil {
			return "", err
		} else if _, ok := values[param.Name]; ok {
			return "", fmt.Errorf("Prototype '%s' has multiple parameters with name '%s'", proto.Name, param.Name)
		}

		quoted, err := param.Quote(val)
		if err != nil {
			return "", err
		}
		values[param.Name] = quoted
	}

	template, err := proto.Template.Body(templateType)
	if err != nil {
		return "", err
	}

	tm := snippet.Parse(strings.Join(template, "\n"))
	text, err := tm.Evaluate(values)
	if err != nil {
		return "", err
	}

	return text, nil
}

func fundUniquePrototype(query string) (*prototype.SpecificationSchema, error) {
	index := prototype.NewIndex([]*prototype.SpecificationSchema{})

	suffixProtos, err := index.SearchNames(query, prototype.Suffix)
	if err != nil {
		return nil, err
	}

	if len(suffixProtos) == 1 {
		// Success.
		return suffixProtos[0], nil
	} else if len(suffixProtos) > 1 {
		// Ambiguous match.
		names := specNames(suffixProtos)
		return nil, fmt.Errorf("Ambiguous match for '%s':\n%s", query, strings.Join(names, "\n"))
	} else {
		// No matches.
		substrProtos, err := index.SearchNames(query, prototype.Substring)
		if err != nil || len(substrProtos) == 0 {
			return nil, fmt.Errorf("No prototype names matched '%s'", query)
		}

		partialMatches := specNames(substrProtos)
		partials := strings.Join(partialMatches, "\n")
		return nil, fmt.Errorf("No prototype names matched '%s'; a list of partial matches:\n%s", query, partials)
	}
}

func specNames(protos []*prototype.SpecificationSchema) []string {
	partialMatches := []string{}
	for _, proto := range protos {
		partialMatches = append(partialMatches, proto.Name)
	}

	return partialMatches
}
