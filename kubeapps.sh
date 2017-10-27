#!/bin/sh
#
# mock implementation

GCTAG=kubeapps

set -x -e

case "$1" in
    up)
        exec kubecfg update -v --gc-tag=$GCTAG kubeapps.jsonnet
        ;;
    down)
        exec kubecfg delete -v --gc-tag=$GCTAG kubeapps.jsonnet
        ;;
    *)
        echo "Unknown subcommand: $1" >&2
        exit 1
esac
