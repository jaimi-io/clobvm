package e2e_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	runner_sdk "github.com/ava-labs/avalanche-network-runner/client"
	"github.com/ava-labs/avalanche-network-runner/rpcpb"
	"github.com/ava-labs/avalanchego/config"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/jaimi-io/clobvm/genesis"
	"github.com/jaimi-io/hypersdk/crypto"
	"github.com/jaimi-io/hypersdk/rpc"
	hutils "github.com/jaimi-io/hypersdk/utils"
	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

const (
	startAmount = uint64(1000000000000)
	sendAmount  = uint64(5000)

	healthPollInterval = 10 * time.Second
)

func TestE2e(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "tokenvm e2e test suites")
}

var (
	requestTimeout time.Duration

	networkRunnerLogLevel string
	gRPCEp                string
	gRPCGatewayEp         string

	execPath  string
	pluginDir string

	vmGenesisPath    string
	vmConfigPath     string
	subnetConfigPath string
	outputPath       string

	mode string

	logsDir string

	blockchainIDA string
	blockchainIDB string

	trackSubnetsOpt runner_sdk.OpOption
)

func init() {
	flag.DurationVar(
		&requestTimeout,
		"request-timeout",
		120*time.Second,
		"timeout for transaction issuance and confirmation",
	)

	flag.StringVar(
		&networkRunnerLogLevel,
		"network-runner-log-level",
		"info",
		"gRPC server endpoint",
	)

	flag.StringVar(
		&gRPCEp,
		"network-runner-grpc-endpoint",
		"0.0.0.0:8080",
		"gRPC server endpoint",
	)
	flag.StringVar(
		&gRPCGatewayEp,
		"network-runner-grpc-gateway-endpoint",
		"0.0.0.0:8081",
		"gRPC gateway endpoint",
	)

	flag.StringVar(
		&execPath,
		"avalanchego-path",
		"",
		"avalanchego executable path",
	)

	flag.StringVar(
		&pluginDir,
		"avalanchego-plugin-dir",
		"",
		"avalanchego plugin directory",
	)

	flag.StringVar(
		&vmGenesisPath,
		"vm-genesis-path",
		"",
		"VM genesis file path",
	)

	flag.StringVar(
		&vmConfigPath,
		"vm-config-path",
		"",
		"VM configfile path",
	)

	flag.StringVar(
		&subnetConfigPath,
		"subnet-config-path",
		"",
		"Subnet configfile path",
	)

	flag.StringVar(
		&outputPath,
		"output-path",
		"",
		"output YAML path to write local cluster information",
	)

	flag.StringVar(
		&mode,
		"mode",
		"test",
		"'test' to shut down cluster after tests, 'run' to skip tests and only run without shutdown",
	)
}

const (
	modeTest      = "test"
	modeFullTest  = "full-test" // runs state sync
	modeRun       = "run"
	modeRunSingle = "run-single"
)

var anrCli runner_sdk.Client

var _ = ginkgo.BeforeSuite(func() {
	gomega.Expect(mode).Should(gomega.Or(
		gomega.Equal(modeTest),
		gomega.Equal(modeFullTest),
		gomega.Equal(modeRun),
		gomega.Equal(modeRunSingle),
	))
	logLevel, err := logging.ToLevel(networkRunnerLogLevel)
	gomega.Expect(err).Should(gomega.BeNil())
	logFactory := logging.NewFactory(logging.Config{
		DisplayLevel: logLevel,
		LogLevel:     logLevel,
	})
	log, err := logFactory.Make("main")
	gomega.Expect(err).Should(gomega.BeNil())

	anrCli, err = runner_sdk.New(runner_sdk.Config{
		Endpoint:    gRPCEp,
		DialTimeout: 10 * time.Second,
	}, log)
	gomega.Expect(err).Should(gomega.BeNil())

	hutils.Outf(
		"{{green}}sending 'start' with binary path:{{/}} %q (%q)\n",
		execPath,
		"kneDvWXFFZ68XzNjnLftBaJ4xARmSPVC4neR8dyR8ERYiKFfe",
	)

	// Start cluster
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	resp, err := anrCli.Start(
		ctx,
		execPath,
		runner_sdk.WithPluginDir(pluginDir),
		// We don't disable PUT gossip here because the E2E test adds multiple
		// non-validating nodes (which will fall behind).
		runner_sdk.WithGlobalNodeConfig(`{
				"log-display-level":"info",
				"proposervm-use-current-height":true,
				"throttler-inbound-validator-alloc-size":"10737418240",
				"throttler-inbound-at-large-alloc-size":"10737418240",
				"throttler-inbound-node-max-processing-msgs":"100000",
				"throttler-inbound-bandwidth-refill-rate":"1073741824",
				"throttler-inbound-bandwidth-max-burst-size":"1073741824",
				"throttler-inbound-cpu-validator-alloc":"100000",
				"throttler-inbound-disk-validator-alloc":"10737418240000",
				"throttler-outbound-validator-alloc-size":"10737418240",
				"throttler-outbound-at-large-alloc-size":"10737418240",
				"snow-mixed-query-num-push-vdr":"10",
				"consensus-on-accept-gossip-validator-size":"10",
				"consensus-on-accept-gossip-peer-size":"10",
				"network-compression-type":"none",
				"consensus-app-concurrency":"512"
			}`),
	)
	cancel()
	gomega.Expect(err).Should(gomega.BeNil())
	hutils.Outf(
		"{{green}}successfully started cluster:{{/}} %s {{green}}subnets:{{/}} %+v\n",
		resp.ClusterInfo.RootDataDir,
		resp.GetClusterInfo().GetSubnets(),
	)
	logsDir = resp.GetClusterInfo().GetRootDataDir()

	// Name 10 new validators (which should have BLS key registered)
	subnetA := []string{}
	subnetB := []string{}
	for i := 1; i <= 10; i++ {
		n := fmt.Sprintf("node%d-bls", i)
		if i <= 5 {
			subnetA = append(subnetA, n)
		} else {
			subnetB = append(subnetB, n)
		}
	}
	specs := []*rpcpb.BlockchainSpec{
		{
			VmName:      "clobvm",
			Genesis:     vmGenesisPath,
			ChainConfig: vmConfigPath,
			SubnetSpec: &rpcpb.SubnetSpec{
				SubnetConfig: subnetConfigPath,
				Participants: subnetA,
			},
		},
		{
			VmName:      "clobvm",
			Genesis:     vmGenesisPath,
			ChainConfig: vmConfigPath,
			SubnetSpec: &rpcpb.SubnetSpec{
				SubnetConfig: subnetConfigPath,
				Participants: subnetB,
			},
		},
	}
	if mode == modeRunSingle {
		specs = specs[0:1]
	}

	// Create 2 subnets
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Minute)
	sresp, err := anrCli.CreateBlockchains(
		ctx,
		specs,
	)
	cancel()
	gomega.Expect(err).Should(gomega.BeNil())

	blockchainIDA = sresp.ChainIds[0]
	subnetIDA := sresp.ClusterInfo.CustomChains[blockchainIDA].SubnetId
	hutils.Outf(
		"{{green}}successfully added chain:{{/}} %s {{green}}subnet:{{/}} %s {{green}}participants:{{/}} %+v\n",
		blockchainIDA,
		subnetIDA,
		subnetA,
	)

	if mode == modeRunSingle {
		trackSubnetsOpt = runner_sdk.WithGlobalNodeConfig(fmt.Sprintf(`{"%s":"%s"}`,
			config.TrackSubnetsKey,
			subnetIDA,
		))
	} else {
		blockchainIDB = sresp.ChainIds[1]
		subnetIDB := sresp.ClusterInfo.CustomChains[blockchainIDB].SubnetId
		hutils.Outf(
			"{{green}}successfully added chain:{{/}} %s {{green}}subnet:{{/}} %s {{green}}participants:{{/}} %+v\n",
			blockchainIDB,
			subnetIDB,
			subnetB,
		)
		trackSubnetsOpt = runner_sdk.WithGlobalNodeConfig(fmt.Sprintf(`{"%s":"%s,%s"}`,
			config.TrackSubnetsKey,
			subnetIDA,
			subnetIDB,
		))
	}

	gomega.Expect(blockchainIDA).Should(gomega.Not(gomega.BeEmpty()))
	if mode != modeRunSingle {
		gomega.Expect(blockchainIDB).Should(gomega.Not(gomega.BeEmpty()))
	}
	gomega.Expect(logsDir).Should(gomega.Not(gomega.BeEmpty()))

	cctx, ccancel := context.WithTimeout(context.Background(), 2*time.Minute)
	status, err := anrCli.Status(cctx)
	ccancel()
	gomega.Expect(err).Should(gomega.BeNil())
	nodeInfos := status.GetClusterInfo().GetNodeInfos()

	instancesA = []instance{}
	var output string
	for _, nodeName := range subnetA {
		info := nodeInfos[nodeName]
		u := fmt.Sprintf("%s/ext/bc/%s", info.Uri, blockchainIDA)
		output += fmt.Sprintf("%s\n", u)
		_, err := ids.FromString(blockchainIDA)
		gomega.Expect(err).Should(gomega.BeNil())
		nodeID, err := ids.NodeIDFromString(info.GetId())
		gomega.Expect(err).Should(gomega.BeNil())
		instancesA = append(instancesA, instance{
			nodeID: nodeID,
			uri:    u,
			cli:    rpc.NewJSONRPCClient(u),
		})
	}

  _ = os.WriteFile("./.uri", []byte(output), 0644)

	if mode != modeRunSingle {
		instancesB = []instance{}
		for _, nodeName := range subnetB {
			info := nodeInfos[nodeName]
			u := fmt.Sprintf("%s/ext/bc/%s", info.Uri, blockchainIDB)
			_, err := ids.FromString(blockchainIDB)
			gomega.Expect(err).Should(gomega.BeNil())
			nodeID, err := ids.NodeIDFromString(info.GetId())
			gomega.Expect(err).Should(gomega.BeNil())
			instancesB = append(instancesB, instance{
				nodeID: nodeID,
				uri:    u,
				cli:    rpc.NewJSONRPCClient(u),
			})
		}
	}

	// Load default pk
	priv, err = crypto.HexToKey(
		"323b1d8f4eed5f0da9da93071b034f2dce9d2d22692c172f3cb252a64ddfafd01b057de320297c29ad0c1f589ea216869cf1938d88c9fbd70d6748323dbf2fa7", //nolint:lll
	)
	gomega.Ω(err).Should(gomega.BeNil())
	rsender = priv.PublicKey()
	hutils.Outf("\n{{yellow}}$ loaded address:{{/}} %s\n\n", rsender)
})

var (
	priv    crypto.PrivateKey
	rsender crypto.PublicKey

	instancesA []instance
	instancesB []instance

	gen *genesis.Genesis
)

type instance struct {
	nodeID ids.NodeID
	uri    string
	cli    *rpc.JSONRPCClient
}

var _ = ginkgo.AfterSuite(func() {
	switch mode {
	case modeTest, modeFullTest:
		hutils.Outf("{{red}}shutting down cluster{{/}}\n")
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		_, err := anrCli.Stop(ctx)
		cancel()
		gomega.Expect(err).Should(gomega.BeNil())

	case modeRun:
		hutils.Outf("{{yellow}}skipping cluster shutdown{{/}}\n\n")
		hutils.Outf("{{cyan}}Blockchain A:{{/}} %s\n", blockchainIDA)
		for _, member := range instancesA {
			hutils.Outf("%s URI: %s\n", member.nodeID, member.uri)
		}
		hutils.Outf("\n{{cyan}}Blockchain B:{{/}} %s\n", blockchainIDB)
		for _, member := range instancesB {
			hutils.Outf("%s URI: %s\n", member.nodeID, member.uri)
		}

	case modeRunSingle:
		hutils.Outf("{{yellow}}skipping cluster shutdown{{/}}\n\n")
		hutils.Outf("{{cyan}}Blockchain:{{/}} %s\n", blockchainIDA)
		for _, member := range instancesA {
			hutils.Outf("%s URI: %s\n", member.nodeID, member.uri)
		}
	}
	gomega.Expect(anrCli.Close()).Should(gomega.BeNil())
})

var _ = ginkgo.Describe("[Ping]", func() {
	ginkgo.It("can ping A", func() {
		for _, inst := range instancesA {
			cli := inst.cli
			ok, err := cli.Ping(context.Background())
			gomega.Ω(ok).Should(gomega.BeTrue())
			gomega.Ω(err).Should(gomega.BeNil())
		}
	})

	ginkgo.It("can ping B", func() {
		for _, inst := range instancesB {
			cli := inst.cli
			ok, err := cli.Ping(context.Background())
			gomega.Ω(ok).Should(gomega.BeTrue())
			gomega.Ω(err).Should(gomega.BeNil())
		}
	})
})

var _ = ginkgo.Describe("[Network]", func() {
	ginkgo.It("can get network A", func() {
		for _, inst := range instancesA {
			cli := inst.cli
			networkID, _, chainID, err := cli.Network(context.Background())
			gomega.Ω(networkID).Should(gomega.Equal(uint32(1337)))
			gomega.Ω(chainID).ShouldNot(gomega.Equal(ids.Empty))
			gomega.Ω(err).Should(gomega.BeNil())
		}
	})

	ginkgo.It("can get network B", func() {
		for _, inst := range instancesB {
			cli := inst.cli
			networkID, _, chainID, err := cli.Network(context.Background())
			gomega.Ω(networkID).Should(gomega.Equal(uint32(1337)))
			gomega.Ω(chainID).ShouldNot(gomega.Equal(ids.Empty))
			gomega.Ω(err).Should(gomega.BeNil())
		}
	})
})