# kubecfg

[![Build Status](https://travis-ci.org/ksonnet/kubecfg.svg?branch=master)](https://travis-ci.org/ksonnet/kubecfg)
[![Go Report Card](https://goreportcard.com/badge/github.com/ksonnet/kubecfg)](https://goreportcard.com/report/github.com/ksonnet/kubecfg)

A tool for managing Kubernetes resources as code.

`kubecfg` allows you to express the patterns across your
infrastructure and reuse these powerful "templates" across many
services, and then manage those templates as files in version control.
The more complex your infrastructure is, the more you will gain from
using kubecfg.

Status: Basic functionality works, and the tool is usable.  The focus
now is on clearer error reporting and advanced features.

Yes, Google employees will recognise this as being very similar to a
similarly-named internal tool ;)

## Install

Pre-compiled executables exist for some platforms on
the [Github releases](https://github.com/ksonnet/kubecfg/releases)
page.

To build from source:

```console
% PATH=$PATH:$GOPATH/bin
% go get github.com/ksonnet/kubecfg
```

Requires golang >=1.7 and a functional cgo environment (C++ with libstdc++).
Note that recent OSX environments
[require golang >=1.8.1](https://github.com/golang/go/issues/19734) to
avoid an immediate `Killed: 9`.

## Quickstart

```console
# Include <kubecfg.git>/lib in kubecfg/jsonnet library search path.
# Can also use explicit `-J` args everywhere.
% export KUBECFG_JPATH=/path/to/kubecfg/lib

# Show generated YAML
% kubecfg show -o yaml -f examples/guestbook.jsonnet

# Create resources
% kubecfg apply -f examples/guestbook.jsonnet

# Modify configuration (downgrade gb-frontend image)
% sed -i.bak '\,gcr.io/google-samples/gb-frontend,s/:v4/:v3/' examples/guestbook.jsonnet
# See differences vs server
% kubecfg diff -f examples/guestbook.jsonnet

# Update to new config
% kubecfg apply -f examples/guestbook.jsonnet

# Clean up after demo
% kubecfg delete -f examples/guestbook.jsonnet
```

## Features

- Supports JSON, YAML or jsonnet files (by file suffix).
- Best-effort sorts objects before updating, so that dependencies are
  pushed to the server before objects that refer to them.
- Additional jsonnet builtin functions. See `lib/kubecfg.libsonnet`.

## Infrastructure-as-code Philosophy

The idea is to describe *as much as possible* about your configuration
as files in version control (eg: git).

Changes to the configuration follow a regular review, approve, merge,
etc code change workflow (github pull-requests, phabricator diffs,
etc).  At any point, the config in version control captures the entire
desired-state, so the system can be easily recreated in a QA cluster
or to recover from disaster.

### Jsonnet

Kubecfg relies heavily on [jsonnet](http://jsonnet.org/) to describe
Kubernetes resources, and is really just a thin Kubernetes-specific
wrapper around jsonnet evaluation.  You should read the jsonnet
[tutorial](http://jsonnet.org/docs/tutorial.html), and skim the functions available in the jsonnet [`std`](http://jsonnet.org/docs/stdlib.html)
library.
