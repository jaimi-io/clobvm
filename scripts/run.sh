#!/usr/bin/env bash
# Copyright (C) 2023, Ava Labs, Inc. All rights reserved.
# See the file LICENSE for licensing terms.

set -e

# to run E2E tests (terminates cluster afterwards)
# MODE=test ./scripts/run.sh
if ! [[ "$0" =~ scripts/run.sh ]]; then
  echo "must be run from repository root"
  exit 255
fi

VERSION=1.10.1
MODE=${MODE:-run-single}
LOGLEVEL=${LOGLEVEL:-info}
AVALANCHE_LOG_LEVEL=${AVALANCHE_LOG_LEVEL:-INFO}
STATESYNC_DELAY=${STATESYNC_DELAY:-0}
PROPOSER_MIN_BLOCK_DELAY=${PROPOSER_MIN_BLOCK_DELAY:-0}
if [[ ${MODE} != "run" && ${MODE} != "run-single" ]]; then
  STATESYNC_DELAY=500000000 # 500ms
  PROPOSER_MIN_BLOCK_DELAY=100000000 # 100ms
fi

echo "Running with:"
echo VERSION: ${VERSION}
echo MODE: ${MODE}
echo STATESYNC_DELAY: ${STATESYNC_DELAY}
echo PROPOSER_MIN_BLOCK_DELAY: ${PROPOSER_MIN_BLOCK_DELAY}

############################
# build avalanchego
# https://github.com/ava-labs/avalanchego/releases
GOARCH=$(go env GOARCH)
GOOS=$(go env GOOS)
AVALANCHEGO_PATH=/tmp/avalanchego-v${VERSION}/avalanchego
AVALANCHEGO_PLUGIN_DIR=/tmp/avalanchego-v${VERSION}/plugins

if [ ! -f "$AVALANCHEGO_PATH" ]; then
  echo "building avalanchego"
  CWD=$(pwd)

  # Clear old folders
  rm -rf /tmp/avalanchego-v${VERSION}
  mkdir -p /tmp/avalanchego-v${VERSION}
  rm -rf /tmp/avalanchego-src
  mkdir -p /tmp/avalanchego-src

  # Download src
  cd /tmp/avalanchego-src
  git clone https://github.com/ava-labs/avalanchego.git
  cd avalanchego
  git checkout v${VERSION}

  # Build avalanchego
  ./scripts/build.sh
  mv build/avalanchego /tmp/avalanchego-v${VERSION}

  cd ${CWD}
else
  echo "using previously built avalanchego"
fi

############################

############################
echo "building clobvm"

# delete previous (if exists)
rm -f /tmp/avalanchego-v${VERSION}/plugins/kneDvWXFFZ68XzNjnLftBaJ4xARmSPVC4neR8dyR8ERYiKFfe

# rebuild with latest code
go build \
-o /tmp/avalanchego-v${VERSION}/plugins/kneDvWXFFZ68XzNjnLftBaJ4xARmSPVC4neR8dyR8ERYiKFfe \
./cmd/clobvm

echo "building clob-cli"
go build -v -o /tmp/clob-cli ./cmd/clob-cli

# log everything in the avalanchego directory
find /tmp/avalanchego-v${VERSION}

############################

############################

# Always create allocations (linter doesn't like tab)
# echo "creating allocations file"
# cat <<EOF > /tmp/allocations.json
# [{"address":"clob1rvzhmceq997zntgvravfagsks6w0ryud3rylh4cdvayry0dl97nsjzf3yp", "balance":1000000000000}]
# EOF

# GENESIS_PATH=$2
# if [[ -z "${GENESIS_PATH}" ]]; then
#   echo "creating VM genesis file with allocations"
#   rm -f /tmp/clobvm.genesis
#   /tmp/clob-cli genesis generate /tmp/allocations.json \
#   --max-block-units 4000000 \
#   --window-target-units 100000000000 \
#   --window-target-blocks 30 \
#   --genesis-file /tmp/clobvm.genesis
# else
#   echo "copying custom genesis file"
#   rm -f /tmp/clobvm.genesis
#   cp ${GENESIS_PATH} /tmp/clobvm.genesis
# fi

############################

echo "creating vm genesis"
rm -f /tmp/clobvm.genesis
cat <<EOF > /tmp/clobvm.genesis
{"hrp":"clob","maxBlockTxs":20000,"maxBlockUnits":18446744073709551615,"baseUnits":0,"validityWindow":60,"minUnitPrice":1,"unitPriceChangeDenominator":48,"windowTargetUnits":20000000000,"minBlockCost":0,"blockCostChangeDenominator":48,"windowTargetBlocks":1000000000}
EOF

############################

echo "creating vm config"
rm -f /tmp/clobvm.config
rm -rf /tmp/clobvm-e2e-profiles
cat <<EOF > /tmp/clobvm.config
{
  "mempoolSize": 10000000,
  "mempoolPayerSize": 10000000,
  "mempoolExemptPayers":["clob1rvzhmceq997zntgvravfagsks6w0ryud3rylh4cdvayry0dl97nsjzf3yp"],
  "parallelism": 5,
  "streamingBacklogSize": 10000000,
  "gossipMaxSize": 32768,
  "gossipProposerDepth": 1,
  "buildProposerDiff": 1,
  "verifyTimeout": 5,
  "trackedPairs":["*"],
  "preferredBlocksPerSecond": 3,
  "continuousProfilerDir":"/tmp/clobvm-e2e-profiles/*",
  "logLevel": "${LOGLEVEL}",
  "stateSyncServerDelay": ${STATESYNC_DELAY}
}
EOF
mkdir -p /tmp/clobvm-e2e-profiles

############################

############################

echo "creating subnet config"
rm -f /tmp/clobvm.subnet
cat <<EOF > /tmp/clobvm.subnet
{
  "proposerMinBlockDelay": ${PROPOSER_MIN_BLOCK_DELAY}
}
EOF

############################

############################
echo "building e2e.test"
# to install the ginkgo binary (required for test build and run)
go install -v github.com/onsi/ginkgo/v2/ginkgo@v2.8.1

# alert the user if they do not have $GOPATH properly configured
if ! command -v ginkgo &> /dev/null
then
    echo -e "\033[0;31myour golang environment is misconfigued...please ensure the golang bin folder is in your PATH\033[0m"
    echo -e "\033[0;31myou can set this for the current terminal session by running \"export PATH=\$PATH:\$(go env GOPATH)/bin\"\033[0m"
    exit
fi

ACK_GINKGO_RC=true ginkgo build ./tests/e2e
./tests/e2e/e2e.test --help

#################################
# download avalanche-network-runner
# https://github.com/ava-labs/avalanche-network-runner
ANR_REPO_PATH=github.com/ava-labs/avalanche-network-runner
ANR_VERSION=v1.4.1
# version set
go install -v ${ANR_REPO_PATH}@${ANR_VERSION}

#################################
# run "avalanche-network-runner" server
GOPATH=$(go env GOPATH)
if [[ -z ${GOBIN+x} ]]; then
  # no gobin set
  BIN=${GOPATH}/bin/avalanche-network-runner
else
  # gobin set
  BIN=${GOBIN}/avalanche-network-runner
fi

killall avalanche-network-runner || true

echo "launch avalanche-network-runner in the background"
$BIN server \
--log-level verbo \
--port=":12352" \
--grpc-gateway-port=":12353" &
PID=${!}

############################
# By default, it runs all e2e test cases!
# Use "--ginkgo.skip" to skip tests.
# Use "--ginkgo.focus" to select tests.

KEEPALIVE=false
function cleanup() {
  if [[ ${KEEPALIVE} = true ]]; then
    echo "avalanche-network-runner is running in the background..."
    echo ""
    echo "use the following command to terminate:"
    echo ""
    echo "killall avalanche-network-runner"
    echo ""
    exit
  fi

  echo "avalanche-network-runner shutting down..."
  killall avalanche-network-runner
}
trap cleanup EXIT

echo "running e2e tests"
./tests/e2e/e2e.test \
--ginkgo.v \
--network-runner-log-level verbo \
--network-runner-grpc-endpoint="0.0.0.0:12352" \
--network-runner-grpc-gateway-endpoint="0.0.0.0:12353" \
--avalanchego-path=${AVALANCHEGO_PATH} \
--avalanchego-plugin-dir=${AVALANCHEGO_PLUGIN_DIR} \
--vm-genesis-path=/tmp/clobvm.genesis \
--vm-config-path=/tmp/clobvm.config \
--subnet-config-path=/tmp/clobvm.subnet \
--output-path=/tmp/avalanchego-v${VERSION}/output.yaml \
--mode=${MODE}

############################
if [[ ${MODE} == "run" || ${MODE} == "run-single" ]]; then
  echo "cluster is ready!"
  # We made it past initialization and should avoid shutting down the network
  KEEPALIVE=true
fi
