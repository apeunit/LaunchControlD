# LaunchControlD

The command & control server for the LaunchControl project. LaunchControl:

1. takes a description of a Cosmos chain's genesis file
2. provisions one virtual machine per Cosmos validator node
3. generates the configuration files for each validator node
4. deploys the Cosmos chain on each VM
5. deploys an instance of the Cosmos light client
6. and optionally a faucet.

This simplifies the task of spinning up chains.

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

default_payload_location:
  docker_image: "apeunit/launchpayload:1.0.0"
  binary_url: https://github.com/apeunit/LaunchPayload/releases/download/v0.0.0/launchpayload-v0.0.0.zip
  binary_path: "/tmp/workspace/bin"
  daemon_path: "/tmp/workspace/bin/launchpayloadd"
  cli_path: "/tmp/workspace/bin/launchpayloadcli"


```

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

Example: To create a new event run the command

```sh
> lctrld events new eventsample1.yml
┌─┐┬  ┬┌┬┐┬  ┬┌─┐╔╦╗
├┤ └┐┌┘ │ └┐┌┘┌─┘ ║║
└─┘ └┘  ┴  └┘ └─┘═╩╝ vv1.1.0-12-g73ac2f7
Using config file: /home/shinichi/source/work/LaunchControlD/config.yaml
Summary:
Validator alice@apeunit.com has initial balance of 500drop,1000000evtx,100000000stake
Including other accounts, the genesis account state is:
alice@apeunit.com: &{Name:alice@apeunit.com Address: Mnemonic: GenesisBalance:500drop,1000000evtx,100000000stake Validator:true Faucet:false ConfigLocation:{CLIConfigDir: DaemonConfigDir:}}
dropgiver: &{Name:dropgiver Address: Mnemonic: GenesisBalance:10000000000drop,10000000000evtx Validator:false Faucet:true ConfigLocation:{CLIConfigDir: DaemonConfigDir:}}
Finally will be deploying 1 servers+nodes (1 for each validators) on virtualbox
Shall we proceed? [Y/n]:
Here we go!!
INFO[0006] alice@apeunit.com's node ID is drop-ed9e103f3f27564342af-0
INFO[0074] Your event ID is drop-ed9e103f3f27564342af
Operation completed in 1m14.70277799s

```
This will start as many virtual machines as there were validators specified in the config.yaml, one instance for each validator.

To list the available events and the status of their nodes run:

```sh
> lctrld events list --verbose
┌─┐┬  ┬┌┬┐┬  ┬┌─┐╔╦╗
├┤ └┐┌┘ │ └┐┌┘┌─┘ ║║
└─┘ └┘  ┴  └┘ └─┘═╩╝ vv1.1.0-12-g73ac2f7
Using config file: /home/shinichi/source/work/LaunchControlD/config.yaml
List events
Event drop-c34efbd55083665002d2 owner: owner@email.com with 1 validators
drop-c34efbd55083665002d2-0 status: Running
drop-c34efbd55083665002d2-0 IP: 188.34.156.184
Event drop-ed9e103f3f27564342af owner: whosaidblockchainisdecentralized@email.com with 1 validators
drop-ed9e103f3f27564342af-0 status: Running
drop-ed9e103f3f27564342af-0 IP: 192.168.99.145
Operation completed in 1.739816533s

```

Now you should setup the payload (Cosmos-SDK based chain) that will run on the machines. The generated config files are stored in the same directory as the event information, under`nodeconfig/`, e.g. `/tmp/workspace/evts/drop-ed9e103f3f27564342af/nodeconfig/`

```sh
> lctrld payload setup $EVTID
┌─┐┬  ┬┌┬┐┬  ┬┌─┐╔╦╗
├┤ └┐┌┘ │ └┐┌┘┌─┘ ║║
└─┘ └┘  ┴  └┘ └─┘═╩╝ vv1.1.0-12-g73ac2f7
Using config file: /home/shinichi/source/work/LaunchControlD/config.yaml
INFO[0000] Initializing daemon configs for each node
INFO[0000] Generating keys for validator accounts
INFO[0000] alice@apeunit.com -> cosmos1ac5rcpu8t3erpl6p6hrqx94cyh744ry596ztsd
INFO[0000] Generating keys for non-validator accounts
INFO[0000] dropgiver -> cosmos1hm66u2k4d7vcljs74g8cfckwpv4ule7yufww9p
INFO[0000] Adding accounts to the genesis.json files
INFO[0000] Creating genesis transactions to turn accounts into validators
INFO[0001] Collecting genesis transactions and writing final genesis.json
INFO[0001] Copying node 0's genesis.json to others and setting up p2p.persistent_peers
alice@apeunit.com's node is 5a8a956afc8d4a614b8ce120e4936b5c8b31d07f@192.168.99.146:26656
INFO[0001] Generating faucet configuration
```

Tell the provisioned machines to run the docker images using the configuration files that were just generated.

```sh
> lctrld payload deploy $EVTID
┌─┐┬  ┬┌┬┐┬  ┬┌─┐╔╦╗
├┤ └┐┌┘ │ └┐┌┘┌─┘ ║║
└─┘ └┘  ┴  └┘ └─┘═╩╝ vv1.1.0-12-g73ac2f7
Using config file: /home/shinichi/source/work/LaunchControlD/config.yaml
INFO[0000] Copying node configs to each provisioned machine
INFO[0001] Running docker pull apeunit/launchpayload:latest on each provisioned machine
INFO[0022] Running the dockerized Cosmos daemons on the provisioned machines
INFO[0023] Running the CLI to provide the Light Client Daemon
INFO[0023] Copying the faucet account and configuration to the first validator machine
INFO[0023] Starting the faucet
```

To stop and remove all the machines and their associated configuration, run
```sh
> lctrld events teardown $EVTID
┌─┐┬  ┬┌┬┐┬  ┬┌─┐╔╦╗
├┤ └┐┌┘ │ └┐┌┘┌─┘ ║║
└─┘ └┘  ┴  └┘ └─┘═╩╝ vv1.1.0-12-g73ac2f7
Using config file: /home/shinichi/source/work/LaunchControlD/config.yaml
Teardown Event
Event ID is drop-ed9e103f3f27564342af
INFO[0000] alice@apeunit.com's node ID is drop-ed9e103f3f27564342af-0
drop-ed9e103f3f27564342af-0 stop: Stopping "drop-ed9e103f3f27564342af-0"...
Machine "drop-ed9e103f3f27564342af-0" was stopped.
drop-ed9e103f3f27564342af-0 rm: About to remove drop-ed9e103f3f27564342af-0
WARNING: This action will delete both local reference and remote instance.
Are you sure? (y/n):
Operation completed in 6.778189622s

```

# Troubleshooting

Here are some common errors that you may encounter in while running the `LaunchControlD` and also how to fix them.

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
└─┘ └┘  ┴  └┘ └─┘═╩╝ vv1.1.0-12-g73ac2f7
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
