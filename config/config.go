package config

import (
	"time"

	"github.com/ava-labs/avalanchego/utils/profiler"
	"github.com/jaimi-io/hypersdk/trace"
)

type Config struct {}

func (c *Config) GetTraceConfig() *trace.Config {
	return nil
}

func (c *Config) GetParallelism() int {
	return 5
}

func (c *Config) GetMempoolSize() int {
	return 10000000
}

func (c *Config) GetMempoolPayerSize() int {
	return 10000000
}

func (c *Config) GetMempoolExemptPayers() [][]byte {
	return nil
}

func (c *Config) GetMempoolVerifyBalances() bool {
	return false
}

func (c *Config) GetStreamingBacklogSize() int {
	return 10000000
}

func (c *Config) GetStateHistoryLength() int {
	return 0
}

func (c *Config) GetStateCacheSize() int {
	return 0
}

func (c *Config) GetAcceptorSize() int {
	return 0
}

func (c *Config) GetStateSyncParallelism() int {
	return 0
}

func (c *Config) GetStateSyncMinBlocks() uint64 {
	return 0
}

func (c *Config) GetStateSyncServerDelay() time.Duration {
	return 0
}

func (c *Config) GetBlockLRUSize() int {
	return 0
}

func (c *Config) GetContinuousProfilerConfig() *profiler.Config {
	return nil
}