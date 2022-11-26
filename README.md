# Dvnet: Docker Virtual Network Plugin
This repository contains the implementation of a compliant *Docker Network Plugin*
capable of instantiating arbitrary virtual networks.

Virtual networks can be defined through JSON files in which one can specify:

1. Whether to allow the containers in the virtual network outbound Internet access.
2. Whether to automatically route the entire network.
3. An arbitrary number of subnets, each with its own CIDR block.
4. An arbitrary number of hosts on each subnet:
    - Each host is capable of running a specific (and possibly different) Docker image.
5. An arbitrary number of routers bridging the different subnets:
    - Each router is capable of running a specific (and possibly different) Docker image.
    - Each router can instantiate a series of `iptables`-based firewall rules.

You can find an example configuration file over at [`demos/quagga/net.json`](demos/quagga/net.json).

The main purpose of this tool is offering fully emulated network topologies allowing for
the study and comprehension of network-related concepts. One can easily:

- Capture traffic in **any** point in the network for online or offline analysis using
  tools such as [`tcpdump(1)`](https://man7.org/linux/man-pages/man1/tcpdump.1.html)
  and [WireShark](https://www.wireshark.org).

- Check whether a given firewall policy implemented with
  [`iptables(8)`](https://man7.org/linux/man-pages/man8/iptables.8.html) accomplishes
  its objective by injecting specific traffic to the router in question.

- Configure and deploy network configuration protocolos such as
  [OSPF](https://www.rfc-editor.org/rfc/rfc1131.pdf)
  and [RIPng](https://datatracker.ietf.org/doc/html/rfc2080) through tools such as
  [Quagga](https://www.nongnu.org/quagga/)

The above are just a taste of what can be done: we're sure you can come up with more use cases!

## Installation
`Dvnet` is a plugin for the Docker daemon. This means **Docker must be installed** for `dvnet` to be
of any use. You can check how to install it for your system by following the
[official documentation](https://docs.docker.com/engine/).

Given the technologies used "under the hood" this plugin will **only** run on Linux-based systems.
We provide statically compiled releases to be used "as-is" on Linux machines running on 64-bit
hardware. You can find those over at the [releases page](https://github.com/pcolladosoto/dvnet/releases).

The installation is just a matter of downloading the executable and then running it. We also provide
a SystemD unit file for convenience: [`dvnet.service`](dvnet.service). You can just copy it to
`/etc/systemd/system/dvnet` and you should be good to go.

In order to ease up the process, we also provide a simple (yet working) automation script: [`install.sh`](install.sh).
You can use it directly from the repo without having to download anything:

    # If you use curl(1)
    $ curl -sSL https://raw.githubusercontent.com/pcolladosoto/dvnet/main/install.sh | sudo bash

    # If you use wget(1)
    $ wget -q -O - https://raw.githubusercontent.com/pcolladosoto/dvnet/main/install.sh | sudo bash

Just like the installer explains, you'll need to manually run the following when you want to start the plugin.
Bear in mind the command will need to be executed with elevated privileges (i.e. `sudo`).

    $ systemctl start dvnet

## Creating and removing a network
As `dvnet` is a compliant Docker plugin, you won't be interacting with it directly. You will just have to issue
`docker network ...` commands to the Docker daemon which will then be relayed to the plugin itself.

Creating a network is matter of running:

    $ docker network create --driver dvnet --opt net.dvnet.def=/path/to/network/definition network-name

We just need to pass a single option: the **absolute** path of a valid JSON network definition. We'll also
have to name the network: this is the name the Docker daemon will refer to this network as.

Network definitions might be arbitrarily complex. What's more, the address assignment on each subnet is **implicit**,
which means you'll know the CIDR block assigned to a particular host, but not necessarily the specific IPv4 address.
You might also want to check whether your network definition is the one you actually intended to define. The best
way to check that is to take a look at the links that have been taken into account when instantiating the network.

All the above information is dumped into a couple of files. Assuming your network definition is contained on
`netDef.json`, you'll see that after creating the network through `docker network create ...` the following two files
will appear in your working (i.e. current) directory:

1. `netDef.ipaddr`: This file is a JSON document containing the IPv4 addresses assigned to each host and router. Bear
   in mind that as routers belong to several subnets, they'll be assigned one address per subnet: it's okay for them
   to appear more than once.

2. `netDef.netg`: This file contains the *edges* (i.e. *links*) in the graph representing the instantiated network. Each
   line contains an initial node name, followed by the nodes they have links to and the number of links. The number of
   links will always be `1`, as these networks aren't modelled as [multigraphs](https://en.wikipedia.org/wiki/Multigraph).
   The important bit is checking each node has links to each of the nodes we expect them to be connected to, according to
   the initial network definition.

These files can be deleted at will: they're not needed at all, they just fulfill an informational purpose. Be sure to
check them if you find yourself wondering things such as: what IPv4 address did `foo` have?

After it's brought up, you can check the network exists with:

    $ docker network ls

You can also take a look to find how the different hosts and routers present in the network definition are
now running containers with:

    $ docker ps

When you're done, you can tear the whole thing up with:

    $ docker network rm network-name

This will leave your machine in the same state it was before the network was instantiated configuration-wise. When
the network is up and running, you can use the familiar `docker exec ...` and `docker cp ...` commands to work
with the containers as if they were regular machines.

## Our default Docker images
In order to mimic regular machines, we have written a couple of `Dockerfiles` (you can check them over at
[`dockerfiles`](dockerfiles)) which just add some additional goodies on top of regular Ubuntu images. The
catch is we're running an [`sshd(8)`](https://man7.org/linux/man-pages/man8/sshd.8.html) daemon as each
container's main process (i.e. the process whose `PID` is `1`). This makes containers run "forever" so that
we can open up shells within them at will. As they are running an SSH daemon we can also
[`ssh(1)`](https://man7.org/linux/man-pages/man1/ssh.1.html) into them without a problem. Bear in mind the
user will always be `root` and that the password is `1234`.

We have uploaded these images to [Docker Hub](https://hub.docker.com). You can use them yourself by specifying
[`pcollado/dhost`](https://hub.docker.com/r/pcollado/dhost) and  [`pcollado/drouter`](https://hub.docker.com/r/pcollado/drouter)
for the hosts and routers, respectively.

These images are just an example of what you can run to serve as hosts and routers in the network. You can
also develop your own images and have at it: you just need to tweak your network definition to use them!

## Uninstallation
Uninstalling `dvnet` is just a matter of stopping it before removing a couple of files. Just like before, we provide
an uninstallation script: [`uninstall.sh`](uninstall.sh).

In order to run it from the repo without having to download anything you can follow the same strategy:

    # If you use curl(1)
    $ curl -sSL https://raw.githubusercontent.com/pcolladosoto/dvnet/main/install.sh | sudo bash

    # If you use wget(1)
    $ wget -q -O - https://raw.githubusercontent.com/pcolladosoto/dvnet/main/install.sh | sudo bash
