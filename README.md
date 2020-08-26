# LaunchControlD

The command & control server for the LaunchControl project

## Configuration

The LaunchControlD project **requires** a configuration file to run properly, here is an example `config.yaml` file:

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
> lctrld events new \
--provider hetzner \
--token EVTX \
--coinbase 10000000 \
--owner owner@emailc.om \
-m validator_n1@email.com \
-m validator_n2@email.com \
-m validator_n3@email.com \
-m validator_n4@email.com \
--stake 10000
```

This will start 4 cx11 instances on hetzner with a node each, one for each validator. Each validator will have a stake of 10k stake tokens. 
The chain will also have an additional token EVTX with a coinbase of 10_000k tokens, each validator account will also own 10_000k/4 EVTX tokens. 

to list the available events and the status of their nodes run:

```
> lctrld events list --verbose
```

# Troubleshooting

Here are some common errors that you may encounter in while running the `LaunchControlD` and also how to fix them.

### Config file not found:

```txt
Error loading config file:  : Config File "config" Not Found in "[/home/andrea/Documents/workspaces/blockchain/eventivize/lctrld/dist /etc/lctrld]"
```

**Cause**: no valid configuration file was found

**Soultion**: create a config file, using the [template above](#configuration)


# References
1. [Create a docker image programmatically](https://docs.docker.com/engine/api/sdk/examples/)
2. [Docker machine](https://docs.docker.com/machine)



