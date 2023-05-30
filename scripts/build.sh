#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

if ! [[ "$0" =~ scripts/build.sh ]]; then
  echo "must be run from repository root"
  exit 255
fi

# Set default binary directory location
name="kneDvWXFFZ68XzNjnLftBaJ4xARmSPVC4neR8dyR8ERYiKFfe"

# Build clobvm, which is run as a subprocess
mkdir -p ./build

echo "Building clobvm in ./build/$name"
go build -o ./build/$name ./cmd/clobvm

echo "Building clob-cli in ./build/clob-cli"
go build -o ./build/clob-cli ./cmd/clob-cli

