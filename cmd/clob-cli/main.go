package main

import (
	"os"

	"github.com/jaimi-io/clobvm/cmd/clob-cli/cmd"
	"github.com/jaimi-io/hypersdk/utils"
)

func main() {
	if err := cmd.Execute(); err != nil {
		utils.Outf("{{red}}clob-cli exited with error:{{/}} %+v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}