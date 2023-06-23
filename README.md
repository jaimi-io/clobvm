# clobvm
ClobVM uses a [HyperSDK fork](https://github.com/jaimi-io/hypersdk) to enable clobvm-specific features such as a custom fee implementation and versioned balance.

## Build & Run

To build and spin up a local instance of a clobvm blockchain (consisting of 5 nodes), run `./scripts/run.sh` in the repo directory. After executing, it will output URIs for the five nodes to the terminal and a `.uri` file.

![Run](https://i.imgur.com/lJN8JmY.png)


## RPC Queries

For state queries, the following JSON body is sent as a POST command to `$NODE_URI/clobapi`

```json
{
    "jsonrpc": "2.0",
    "method": "clobvm.[state query endpoint]",
    "params": {},
    "id": 1
}
```

For example, to get the AVAX balance of `clob12l2xyad754fu3s9rqdwq4mkllnl0vc5yerygn2aw5xasmjhzmtwspkx4ek` the following is sent:

```json
{
    "jsonrpc": "2.0",
    "method": "clobvm.balance",
    "params": {
      "address": "clob12l2xyad754fu3s9rqdwq4mkllnl0vc5yerygn2aw5xasmjhzmtwspkx4ek",
      "tokenID": "VmwmdfVNQLiP1zJWmhaHipksKBAHmDZH5rZvdfCQfQ9peNx8a"
    },
    "id": 1
}
```

## clob-cli Setup
To easily run the CLI ensure a private key is input at file `.key.pk` and the node URIs in `.uri`. The following commands are supported:

![CLI Commands](https://i.imgur.com/TfaLqNz.png)

When executing a spam command, the output is as follows:

![Spam Command](https://i.imgur.com/LUtqiZd.png)

To view metrics during a spam test execute the `clob-cli prometheus` command, which provides a prometheus command to run. Once running it acts as a web server that displays time series data on the state of the blockchain i.e. 

![Prometheus Server](https://i.imgur.com/lotuN1V.png)
