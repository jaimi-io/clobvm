package config

import (
	"runtime"
	"time"

	"github.com/ava-labs/avalanchego/utils/profiler"
	"github.com/jaimi-io/hypersdk/trace"
)

type Config struct {}

func (c *Config) GetTraceConfig() *trace.Config {
	return nil
}

func (c *Config) GetParallelism() int {
	numCPUs := runtime.NumCPU()
	if numCPUs > 4 {
		return numCPUs - 4
	}
	return 1
}

func (c *Config) GetMempoolSize() int {
	return 2_048
}

func (c *Config) GetMempoolPayerSize() int {
	return 32
}

func (c *Config) GetMempoolExemptPayers() [][]byte {
	return nil
}

func (c *Config) GetMempoolVerifyBalances() bool {
	return false
}

func (c *Config) GetStreamingBacklogSize() int {
	return 1024
}

func (c *Config) GetStateHistoryLength() int {
	return 256
}

func (c *Config) GetStateCacheSize() int {
	return 65_536
}

func (c *Config) GetAcceptorSize() int {
	return 1024
}

func (c *Config) GetStateSyncParallelism() int {
	return 4
}

func (c *Config) GetStateSyncMinBlocks() uint64 {
	return 256
}

func (c *Config) GetStateSyncServerDelay() time.Duration {
	return 0
}

func (c *Config) GetBlockLRUSize() int {
	return 128
}

func (c *Config) GetContinuousProfilerConfig() *profiler.Config {
	return &profiler.Config{Enabled: false}
}