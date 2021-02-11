# LaunchControlD

The command & control server for the LaunchControl project. LaunchControl:

1. takes a description of a Cosmos chain's genesis file
2. provisions one virtual machine per Cosmos validator node
3. generates the configuration files for each validator node
4. deploys the Cosmos chain on each VM
5. deploys an instance of the Cosmos light client
6. and optionally a faucet.

This simplifies the task of spinning up chains.

## Installation 

Before you start make sure that you have docker installed:

*lctrld* needs docker to run the docker images to generate configuration files for the nodes. Remember to login to your Docker Hub account to be able to download images.

To provision nodes on your own computer, install virtualbox (docker-machine has a built in virtualbox module). Otherwise lctrld expects to provision nodes using a cloud provider.


### Via go package manager

If you have go installed you can run

```sh
go get github.com/apeunit/LaunchControlD
```

### From source

To install it from source run 

```sh
git clone git@github.com:apeunit/LaunchControlD.git
cd LaunchControlD
make build # binaries now in dist/
make install # installs to GOPATH/bin
cd ..
```

## Check your installation                
    
if everything went well you should able to run
```sh
lctrld --version
lctrld version v1.0.0
```

## Configuration

The LaunchControlD project **requires** a configuration file, which specifies how lctrld can provision and manage virtual machines.

Here is an example `config.yaml` file:

```yaml
---
# the workspace folder will be created if not exists
workspace: "/tmp/workspace"
# this section is docker-machine configuration
docker_machine:
  # add additional folders to the search path while executing the docker-machine command
  search_path:
  - "/usr/local/bin"
  - "/usr/bin"
  # version of the docker-machine release (only for reference)
  version: "0.16.2"
  binary_url: https://github.com/docker/machine/releases/download/v0.16.2/docker-machine-Linux-x86_64
  binary: docker-machine
  env:
  - "MACHINE_DOCKER_INSTALL_URL=https://releases.rancher.com/install-docker/19.03.9.sh"
  - "VIRTUALBOX_BOOT2DOCKER_URL=/your/local/copy/of/boot2docker.iso"

  # drivers for docker machine
  drivers:
    hetzner:
      version: "2.1.0"
      binary: docker-machine-driver-hetzner
      binary_url: https://github.com/JonasProgrammer/docker-machine-driver-hetzner/releases/download/2.1.0/docker-machine-driver-hetzner_2.1.0_linux_amd64.tar.gz
      # in params should be set all the custom parameters for the specific driver
      params:
      - "--hetzner-api-token=xyz"
      # the env vars listed here will be added to the environment
      env:
      - "HETZNER_API_TOKEN=xyz"
    ovh:
      version: "v1.1.0"
      binary_url: https://github.com/yadutaf/docker-machine-driver-ovh/releases/download/v1.1.0-1/docker-machine-driver-ovh-v1.1.0-1-linux-amd64.tar.gz
      binary: docker-machine-driver-ovh
      env:
      - "OVH_APPLICATION_SECRET=abc"
      - "OVH_APPLICATION_KEY=abc"
      - "OVH_CONSUMER_KEY=abc"
    digitalocean:
      # the driver is included, no need to download anything
      env:
      - "DIGITALOCEAN_ACCESS_TOKEN=123"

```

If you will be deploying nodes often using Virtualbox, it is useful to download your own copy of [boot2docker.iso](https://github.com/boot2docker/boot2docker/releases/tag/v19.03.12) so that you won't have to download it every time. Specify the path with `VIRTUALBOX_BOOT2DOCKER_URL`.

Other drivers can be added in the configuration file (a list of available drivers can be found [here](https://github.com/docker/docker.github.io/blob/master/machine/AVAILABLE_DRIVER_PLUGINS.md)).

> 💡: when adding a driver in the drivers section use the name as described in the [official documentation](https://docs.docker.com/machine/drivers/).

## Usage

The first step is to setup the environment using the command

```sh
> lctrld setup [--config config.yaml]
```


> 💡: the default location for the config file is `/etc/lctrld/config.yaml`

This command will setup the environment, download docker-machine and the drivers and create the necessary folders.
It is usually necessary to run the setup only once, but you may have to run it again if you change the configuration,
like for example you add new drivers.

> ⚠️: the workspace path cannot be changed once you have an environment running

### Events

To manage the events lifecycle use the command

```sh
> lctrld events
```

## Example

The following example will create a new event composed by two validator nodes and a faucet running on 3 virtual machine on a local virtualbox installation. It is assumed that `lctrld` is already installed.

**Prerequisites**
- To run the example it is required to install virtualbox on the local machine
- The examples assumes that your working directory is `/lctrld`

First download the sample virtualbox configuration:

```sh
curl -LO https://raw.githubusercontent.com/apeunit/LaunchControlD/master/examples/config_virtualbox.yml
```

Then download the latest version of the `boot2docker` iso image

```
curl -LO https://github.com/boot2docker/boot2docker/releases/download/v19.03.12/boot2docker.iso
```

And finally download the event specification

```sh
curl -LO https://raw.githubusercontent.com/apeunit/LaunchControlD/master/examples/simple_event_w_faucet.yml
```

Now we can run the setup:
```sh
> lctrld setup --config config_virtualbox.yml
```

The output of the setup command should look like this:
```sh
┌─┐┬  ┬┌┬┐┬  ┬┌─┐╔╦╗
├┤ └┐┌┘ │ └┐┌┘┌─┘ ║║
└─┘ └┘  ┴  └┘ └─┘═╩╝ v1.0.0-2-g360d24d
Using config file: config_virtualbox.yml
Setup LaunchControlD started
Setup completed in  1m1.007284248s

```

and the content of the current directory should be

```sh
ls --tree
.
├── boot2docker.iso
├── config_virtualbox.yml
├── simple_event_w_faucet.yml
└── workspace
   ├── bin
   │  └── docker-machine
   ├── evts
   └── tmp
      └── 445467972
```

Once the setup is completed we can setup the event

```sh
> lctrld events new simple_event_w_faucet.yml \
--provider virtualbox \
--config config_virtualbox.yml

┌─┐┬  ┬┌┬┐┬  ┬┌─┐╔╦╗
├┤ └┐┌┘ │ └┐┌┘┌─┘ ║║
└─┘ └┘  ┴  └┘ └─┘═╩╝ v1.0.0-3-ga756f1b
Using config file: config_virtualbox.yml
Summary:
Validator alice@apeunit.com has initial balance of 500drop,1000000evtx,100000000stake
Validator bob@apeunit.com has initial balance of 500drop,1000000evtx,100000000stake
Including other accounts, the genesis account state is:
alice@apeunit.com: &{Name:alice@apeunit.com Address: Mnemonic: GenesisBalance:500drop,1000000evtx,100000000stake Validator:true Faucet:false ConfigLocation:{CLIConfigDir: DaemonConfigDir:}}
bob@apeunit.com: &{Name:bob@apeunit.com Address: Mnemonic: GenesisBalance:500drop,1000000evtx,100000000stake Validator:true Faucet:false ConfigLocation:{CLIConfigDir: DaemonConfigDir:}}
dropgiver: &{Name:dropgiver Address: Mnemonic: GenesisBalance:10000000000drop,10000000000evtx Validator:false Faucet:true ConfigLocation:{CLIConfigDir: DaemonConfigDir:}}
Finally will be deploying 2 servers+nodes (1 for each validators) on virtualbox
Shall we proceed? [Y/n]:Y
Here we go!!
INFO[0001] alice@apeunit.com's node ID is drop-c34efbd55083665002d2-0
INFO[0062] bob@apeunit.com's node ID is drop-c34efbd55083665002d2-1
INFO[0117] Your event ID is drop-c34efbd55083665002d2
Operation completed in 1m57.477250453s

```
This will start as many virtual machines as there were validators specified in the `simple_event_w_faucet.yml`, **one instance for each validator**.

Take note of the event ID (`drop-c34efbd55083665002d2`) since it will be used later

To list the available events and the status of their nodes run:

```sh
> lctrld events list --verbose --config config_virtualbox.yml
┌─┐┬  ┬┌┬┐┬  ┬┌─┐╔╦╗
├┤ └┐┌┘ │ └┐┌┘┌─┘ ║║
└─┘ └┘  ┴  └┘ └─┘═╩╝ v1.0.0-3-ga756f1b
Using config file: config_virtualbox.yml
List events
Event drop-c34efbd55083665002d2 owner: owner@email.com with 2 validators
drop-c34efbd55083665002d2-0 status: Running
drop-c34efbd55083665002d2-0 IP: 192.168.99.108
drop-c34efbd55083665002d2-1 status: Running
drop-c34efbd55083665002d2-1 IP: 192.168.99.109
Operation completed in 1.457770547s
```

At this point we have installed and setup `lctrld`, prepared the configuration files and provisioned the infrastructure for our event network to run on.

What is left to do is to actually deploy the nodes (Cosmos-SDK based) on the network, or as we will refer to it later, to deploy the `payload`.

> 💡: The generated event config files are stored in the same directory as the event ID, under`nodeconfig/`, in this case: `/lctrld/workspace/evts/drop-c34efbd55083665002d2/nodeconfig/`

```sh
> lctrld payload setup drop-c34efbd55083665002d2 --config config_virtualbox.yml

┌─┐┬  ┬┌┬┐┬  ┬┌─┐╔╦╗
├┤ └┐┌┘ │ └┐┌┘┌─┘ ║║
└─┘ └┘  ┴  └┘ └─┘═╩╝ v1.0.0-3-ga756f1b
Using config file: config_virtualbox.yml
INFO[0000] Initializing daemon configs for each node
INFO[0000] Generating keys for validator accounts
INFO[0000] alice@apeunit.com -> cosmos16vj34rzjwlqnuudh0yagsf42xk9c4jxhfzqsh3
INFO[0000] bob@apeunit.com -> cosmos1kgpxkmu8uk7kgj7cka59zsl0wzlyxacqnuv9w4
INFO[0000] Generating keys for non-validator accounts
INFO[0000] dropgiver -> cosmos16kau2asdta7un6cyr08czgunxxzexl56zy0qcd
INFO[0000] Adding accounts to the genesis.json files
INFO[0000] Creating genesis transactions to turn accounts into validators
INFO[0000] Collecting genesis transactions and writing final genesis.json
INFO[0000] Copying node 0's genesis.json to others and setting up p2p.persistent_peers
INFO[0000] otherGenesis: /lctrld/workspace/evts/drop-c34efbd55083665002d2/nodeconfig/1/daemon/config/genesis.json
alice@apeunit.com's node is 9ca0730f9ed0435cde89aaa21233cc202a4b6886@192.168.99.108:26656
bob@apeunit.com's node is f9c467f5d247f4aec2f18936a3ef078f1ab60c9c@192.168.99.109:26656
INFO[0000] Generating faucet configuration

```

Tell the provisioned machines to run the docker images using the configuration files that were just generated.

```sh
> lctrld payload deploy drop-c34efbd55083665002d2 --config config_virtualbox.yml

┌─┐┬  ┬┌┬┐┬  ┬┌─┐╔╦╗
├┤ └┐┌┘ │ └┐┌┘┌─┘ ║║
└─┘ └┘  ┴  └┘ └─┘═╩╝ v1.0.0-3-ga756f1b
Using config file: config_virtualbox.yml
INFO[0000] Copying node configs to each provisioned machine
INFO[0000] Running docker pull apeunit/launchpayload:v1.0.0 on each provisioned machine
INFO[0060] Running the dockerized Cosmos daemons on the provisioned machines
INFO[0061] Running the CLI to provide the Light Client Daemon
INFO[0061] Copying the faucet account and configuration to the first validator machine
INFO[0061] Starting the faucet
```

And it's done! 🎉

You can see it working by pointing your browser to one of the nodes faucet:

```sh
curl -s http://192.168.99.108:8000/status | jq
```
```json
{
  "node_info": {
    "protocol_version": {
      "p2p": "7",
      "block": "10",
      "app": "0"
    },
    "id": "9ca0730f9ed0435cde89aaa21233cc202a4b6886",
    "listen_addr": "tcp://0.0.0.0:26656",
    "network": "drop-c34efbd55083665002d2",
    "version": "0.33.7",
    "channels": "4020212223303800",
    "moniker": "alice@apeunit.com node drop-c34efbd55083665002d2-0",
    "other": {
      "tx_index": "on",
      "rpc_address": "tcp://0.0.0.0:26657"
    }
  },
  "sync_info": {
    "latest_block_hash": "D53320F8DDDF32F5DBB0D448489D1ACDEDA944566118A6EBDB3DE59A2527B9BD",
    "latest_app_hash": "2636C416614501773F86501EFFE3801CDE310BA092BC875E8D547F58C1C68D8E",
    "latest_block_height": "2",
    "latest_block_time": "2021-01-27T10:26:50.848598231Z",
    "earliest_block_hash": "32A9CBFBAC4ED2B79052537F40F6F5E5304D8AFA688F85C37D9743F6BCEA9185",
    "earliest_app_hash": "",
    "earliest_block_height": "1",
    "earliest_block_time": "2021-01-27T10:24:44.152679036Z",
    "catching_up": false
  },
  "validator_info": {
    "address": "F8E33C9DDB3980687A8465CB3D03697E9D227D6B",
    "pub_key": {
      "type": "tendermint/PubKeyEd25519",
      "value": "SHa1pbySh+4D+axdeyBjZmauSkfa3V2eStawjbv59bQ="
    },
    "voting_power": "100"
  }
}
```




To stop and remove all the machines and their associated configuration, run
```sh
> lctrld events teardown drop-c34efbd55083665002d2 --config config_virtualbox.yml

┌─┐┬  ┬┌┬┐┬  ┬┌─┐╔╦╗
├┤ └┐┌┘ │ └┐┌┘┌─┘ ║║
└─┘ └┘  ┴  └┘ └─┘═╩╝ v1.0.0-3-ga756f1b
Using config file: config_virtualbox.yml
Teardown Event
Event ID is drop-c34efbd55083665002d2
INFO[0000] alice@apeunit.com's node ID is drop-c34efbd55083665002d2-0
drop-c34efbd55083665002d2-0 stop: Stopping "drop-c34efbd55083665002d2-0"...
Machine "drop-c34efbd55083665002d2-0" was stopped.
drop-c34efbd55083665002d2-0 rm: About to remove drop-c34efbd55083665002d2-0
WARNING: This action will delete both local reference and remote instance.
Are you sure? (y/n):
INFO[0005] bob@apeunit.com's node ID is drop-c34efbd55083665002d2-1
drop-c34efbd55083665002d2-1 stop: Stopping "drop-c34efbd55083665002d2-1"...
Machine "drop-c34efbd55083665002d2-1" was stopped.
drop-c34efbd55083665002d2-1 rm: About to remove drop-c34efbd55083665002d2-1
WARNING: This action will delete both local reference and remote instance.
Are you sure? (y/n):
Operation completed in 10.148454512s
```

# Troubleshooting

Here are some common errors that you may encounter while running the `LaunchControlD` and also how to fix them.
### Stale directories from errored commands prevent you from retrying the command
Sometimes provisioning a virtual machine with `lctrld events new eventsample1.yml` can error out. When this happens, the event directory e.g. `/tmp/workspace/evts/drop-c34efbd55083665002d2/` is left over. If you remove it, lctrld and docker-machine will think there is no virtual machine, but you may have to remove the virtual machine in Virtualbox using `VBoxManage unregistervm <vm name>`. Or if you are using Hetzner, make sure the VPS is removed, and delete the corresponding SSH key.

In the following example, the previous invocation of `lctrld payload setup drop-c34efbd55083665002d2` failed, so `/tmp/workspace/evts/drop-c34efbd55083665002d2/nodeconfig` was created, but not deleted. `nodeconfig` is where the generated blockchain node configurations and accounts are stored. Simply remove `/tmp/workspace/evts/drop-c34efbd55083665002d2/nodeconfig/` and rerun the command.
```sh
> lctrld payload setup drop-c34efbd55083665002d2

┌─┐┬  ┬┌┬┐┬  ┬┌─┐╔╦╗
├┤ └┐┌┘ │ └┐┌┘┌─┘ ║║
└─┘ └┘  ┴  └┘ └─┘═╩╝ va61f4a0
Using config file: ./config.yaml
INFO[0000] Initializing daemon configs for each node
ERRO[0000] [/tmp/workspace/bin/launchpayloadd init alice@apeunit.com node drop-c34efbd55083665002d2-0 --home /tmp/workspace/evts/drop-c34efbd55083665002d2/nodeconfig/0/daemon --chain-id drop-c34efbd55083665002d2] failed with exit status 1, ERROR: genesis.json file already exists: /tmp/workspace/evts/drop-c34efbd55083665002d2/nodeconfig/0/daemon/config/genesis.json

ERRO[0000] /tmp/workspace/bin/launchpayloadd [/tmp/workspace/bin/launchpayloadd init alice@apeunit.com node drop-c34efbd55083665002d2-0 --home /tmp/workspace/evts/drop-c34efbd55083665002d2/nodeconfig/0/daemon --chain-id drop-c34efbd55083665002d2] failed with exit status 1,
Error: exit status 1
Usage:
  lctrld payload setup EVENTID [flags]

Flags:
  -h, --help   help for setup

Global Flags:
      --config string   config file (default is /etc/lctrld/config.yaml)
  -d, --debug           Enable debug logging

exit status 1
```

### docker-machine errors while provisioning virtual machines
If this happens (especially when docker-machine [installs docker > 19.03](https://github.com/docker/machine/issues/4858)) a basic knowledge of how to use docker-machine will save the day. SSH into the virtual machine and troubleshoot the problem.
```sh
> docker-machine -s /tmp/workspace/evts/drop-c34efbd55083665002d2/.docker/machine ls
NAME                          ACTIVE   DRIVER    STATE     URL                         SWARM   DOCKER     ERRORS
drop-c34efbd55083665002d2-0   -        hetzner   Running   tcp://188.34.156.184:2376           v19.03.9

> docker-machine -s /tmp/workspace/evts/drop-c34efbd55083665002d2/.docker/machine ssh drop-c34efbd55083665002d2-0
Welcome to Ubuntu 18.04.5 LTS (GNU/Linux 4.15.0-126-generic x86_64)

 * Documentation:  https://help.ubuntu.com
 * Management:     https://landscape.canonical.com
 * Support:        https://ubuntu.com/advantage

 * Canonical Livepatch is available for installation.
   - Reduce system reboots and improve kernel security. Activate at:
     https://ubuntu.com/livepatch
Last login: Mon Jan 11 11:54:43 2021 from 95.90.200.92
**root@drop-c34efbd55083665002d2-0:~#**
```

Once you have fixed the problem and the virtual machine's dockerd is listening on `*:2376`, tell `lctrld` to reread the docker-machine state into `/tmp/workspace/evts/drop-c34efbd55083665002d2/event.json`, then continue with the `lctrld payload` subcommands.
```sh
> lctrld events retry drop-c34efbd55083665002d2

┌─┐┬  ┬┌┬┐┬  ┬┌─┐╔╦╗
├┤ └┐┌┘ │ └┐┌┘┌─┘ ║║
└─┘ └┘  ┴  └┘ └─┘═╩╝ v1.0.0
Using config file: /home/shinichi/source/work/LaunchControlD/config.yaml
INFO[0000] Updated info for alice@apeunit.com: &model.MachineConfig{N:"0", EventID:"drop-c34efbd55083665002d2", DriverName:"", TendermintNodeID:"", Instance:model.MachineConfigInstance{IPAddress:"188.34.156.184", MachineName:"drop-c34efbd55083665002d2-0", SSHUser:"root", SSHPort:22, SSHKeyPath:"/tmp/workspace/evts/drop-c34efbd55083665002d2/.docker/machine/machines/drop-c34efbd55083665002d2-0/id_rsa", StorePath:"/tmp/workspace/evts/drop-c34efbd55083665002d2/.docker/machine"}}
```

### Config file not found:

```txt
Error loading config file:  : Config File "config" Not Found in "[/home/andrea/Documents/workspaces/blockchain/eventivize/lctrld/dist /etc/lctrld]"
```

**Cause**: no valid configuration file was found

**Solution**: create a config file, using the [template above](#configuration)


# References
1. [Create a docker image programmatically](https://docs.docker.com/engine/api/sdk/examples/)
2. [Docker machine](https://docs.docker.com/machine)


# REST API

The LaunchControlD provides a set of API to control it's behavior. To run the LaunchControlD API run the command:

```
lctrld serve [--config /path/to/config]
```

once the API are started, the documentation is accessible at the URL:

```
$HOST/swagger/
```

where `$HOST` is the address where you run the API

### Authentication

The API provide a simple authentication mechanism that is token based. To be able to use the API first it is required to register using email/password.

Once registered the API require to make  login call to obtain a temporary token, the
token is exchanged via the header named `X-Lctrld-Token` and it is valid for 12h.
