#!/bin/sh

set -e

if [ ! -f "build/env.sh" ]; then
    echo "$0 must be run from the root of the repository."
    exit 2
fi

# Create fake Go workspace if it doesn't exist yet.
workspace="$PWD/build/_workspace"
root="$PWD"
okcdir="$workspace/src/github.com/okcoin"
if [ ! -L "$okcdir/go-okcoin" ]; then
    mkdir -p "$okcdir"
    cd "$okcdir"
    ln -s ../../../../../. go-okcoin
    cd "$root"
fi

# Set up the environment to use the workspace.
GOPATH="$workspace"
export GOPATH

# Run the command inside the workspace.
cd "$okcdir/go-okcoin"
PWD="$okcdir/go-okcoin"

# Launch the arguments with the configured environment.
exec "$@"
