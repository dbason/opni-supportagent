# opni-supportagent

The Opni support agent will take the controlplane logs collected by the Rancher Log Collector and ingest them into an Opni cluster.  The agent currently supports controlplane logs gathered from RKE, K3s, and RKE2 clusters.

To ingest the logs the CLI must be run from the root director of the unzipped log bundle.

## Usage
  opni-support [flags]
  opni-support [command]

Available Commands:
  completion  generate the autocompletion script for the specified shell
  help        Help about any command
  local       build local opnicluster in k3d and publish logs to it
  publish     publish support bundle to remote opni cluster

Flags:
      --clustername string   cluster name to add as metadata to the logs (default "default")
  -h, --help                 help for opni-support

Use "opni-support [command] --help" for more information about a command.

### local command
This will create a local k3d cluster called opni-support, install opni into the cluster, and then ingest the logs into it.  The k3d binary is not required, however Docker must be installed on the local machine for this to work.

The local command requires one argument, the type of distribution to ingest.  This must be one of rke, k3s, or rke2

### publish command
The publish command works similarly to local, however instead of creating a local cluster this will publish to the payload-receiver endpoint of a remote Opni cluster.

The publish command requires two arguments; the distribution to ingest, and the URL of the payload-receiver endpoint.

## Building the binary locally
The build process uses dapper.  Due to this Docker is required to build the binary.  With docker installed the binaries can be built with the following command:
```bash
CROSS=1 make
```
This will make binaries for linux and darwin in the `./bin/` folder