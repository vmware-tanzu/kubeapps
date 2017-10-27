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
        # This assumes kubeapps.jsonnet is in sync with what's
        # currently running.
        # FIXME(gus): add support for deletion using the garbage
        # collection mechanism.
        exec kubecfg delete -v kubeapps.jsonnet
        ;;
    *)
        echo "Unknown subcommand: $1" >&2
        exit 1
esac
