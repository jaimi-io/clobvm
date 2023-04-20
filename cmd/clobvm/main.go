package main

import (
	"context"
	"fmt"
	"os"

	log "github.com/inconshreveable/log15"

	"github.com/jaimi-io/clobvm/chain"

	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/avalanchego/utils/ulimit"
	"github.com/ava-labs/avalanchego/vms/rpcchainvm"
)

func main() {
	version, err := PrintVersion()
	if err != nil {
		fmt.Printf("couldn't get config: %s", err)
		os.Exit(1)
	}
	// Print VM ID and exit
	if version {
		fmt.Printf("%s@%s\n", chain.Name, chain.Version)
		os.Exit(0)
	}

	if err := ulimit.Set(ulimit.DefaultFDLimit, logging.NoLog{}); err != nil {
		fmt.Printf("failed to set fd limit correctly due to: %s", err)
		os.Exit(1)
	}

	log.Root().SetHandler(log.LvlFilterHandler(log.LvlInfo, log.StreamHandler(os.Stderr, log.TerminalFormat())))

	err = rpcchainvm.Serve(context.Background(), &chain.VM{})
	if err != nil {
		fmt.Printf("failed to serve due to: %s", err)
		os.Exit(1)
	}
}