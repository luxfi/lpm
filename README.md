# Lux Plugin Manager (LPM)

**Note: This code is currently in Alpha. Proceed at your own risk.**

`lpm` is a command-line tool to manage virtual machines binaries for
the [Lux Network](https://github.com/luxfi/node).

`lpm` allows users to build their own custom repositories to provide virtual machine and subnet definitions outside of
the [plugins-core](https://github.com/luxfi/plugins-core) repository. `plugins-core`
is a community-sourced set of plugins and subnets that ships with the `lpm`, but users have the option of adding their own using
the `add-repository` command.

## Installation

### Pre-Built Binaries
#### Instructions

To download a binary for the latest release, run:

```
curl -sSfL https://raw.githubusercontent.com/luxfi/pm/master/scripts/install.sh | sh -s
```

The binary will be installed inside the `./bin` directory (relative to where the install command was run).

_Downloading binaries from the Github UI will cause permission errors on Mac._

To add the binary to your path, run

```
cd bin
export PATH=$PWD:$PATH
```

To add it to your path permanently, add an export command to your shell initialization script (ex: .bashrc).

#### Installing in Custom Location

To download the binary into a specific directory, run:

```
curl -sSfL https://raw.githubusercontent.com/luxfi/pm/master/scripts/install.sh | sh -s -- -b <relative directory>
```

### Source
If you are planning on building from source, you will need [golang](https://go.dev/doc/install) >= 1.18.x installed.

To build from source, you can use the provided build script from the repository root.
```
./scripts/build.sh
```
The resulting `lpm` binary will be available in `./build/lpm`.

## Commands

### add-repository
Starts tracking a plugin repository.

```shell
lpm add-repository --alias luxfi/core --url https://github.com/luxfi/plugins-core.git --branch master
```

#### Parameters:
- `--alias`: The alias of the repository to track (must be in the form of `foo/bar` i.e organization/repository).
- `--url`: The url to the repository.
- `--branch`: The branch name to track.

### install-vm
Installs a virtual machine by its alias. Either a partial alias (e.g `spacesvm`) or a fully qualified name including the repository (e.g `luxfi/core:spacesvm`) to disambiguate between multiple repositories can be used.

If multiple matches are found (e.g `repository-1/foovm`, `repository-2/foovm`), you will be required to specify the
fully qualified name of the virtual machine to disambiguate the repository to install from.

This will install the virtual machine binary to your `node` plugin path.

```shell
lpm install-vm --vm spacesvm
```

#### Parameters:
- `--vm`: The alias of the VM to install.


### join-subnet
Joins a subnet by its alias. Either a partial alias (e.g `spaces`) or a fully qualified name including the repository (e.g `luxfi/core:spaces`) to disambiguate between multiple repositories can be used.

This will install dependencies for the subnet by calling `install-vm` on each virtual machine required by the subnet.

If multiple matches are found (e.g `repository-1/foo`, `repository-2/foo`), you will be required to specify the
fully qualified name of the subnet definition to disambiguate the repository to install from.


```shell
lpm join-subnet --subnet spaces
```

#### Parameters:
- `--subnet`: The alias of the VM to install.

### list-repositories
Lists all tracked repositories.

```shell
lpm list-repositories
```

### uninstall-vm
Installs a virtual machine by its alias.

If multiple matches are found (e.g `repository-1/foovm`, `repository-2/foovm`), you will be required to specify the
fully qualified name of the virtual machine to disambiguate the repository to install from.

This will remove the virtual machine binary from your `node` plugin path.

```shell
lpm uninstall-vm --vm spacesvm
```

#### Parameters:
- `--vm`: The alias of the VM to uninstall.

### update

Fetches the latest plugin definitions from all tracked repositories.


```shell
lpm list-repositories
```

### upgrade

Upgrades a virtual machine binary. If one is not provided, this will upgrade all virtual machine binaries in your
`node` plugin path with the latest synced definitions.

For a virtual machine to be upgraded, it must have been installed using the `lpm`.

```shell
lpm upgrade
```

#### Parameters
- `--vm`: (Optional) The alias of the VM to upgrade. If none is provided, all VMs are upgraded.

### remove-repository
Stops tracking a repository and wipes all local definitions from that repository.

```shell
lpm remove-repository --alias organization/repository
```

#### Parameters:
- `--alias`: The alias of the repository to start tracking.

## Examples

###
1. Install the spaces subnet!
```shell
./build/lpm join-subnet --subnet spaces
```

2. You'll see some output like this:
```text
$ ./build/lpm join-subnet --subnet spaces

Installing virtual machines for subnet Ai42MkKqk8yjXFCpoHXw7rdTWSHiKEMqh5h8gbxwjgkCUfkrk.
Downloading https://github.com/luxfi/spacesvm/archive/refs/tags/v0.0.3.tar.gz...
HTTP response 200 OK
Calculating checksums...
Saw expected checksum value of 1ac250f6c40472f22eaf0616fc8c886078a4eaa9b2b85fbb4fb7783a1db6af3f
Creating sources directory...
Unpacking luxfi/network-plugins-core:spacesvm...
Running install script at scripts/build.sh...
Building spacesvm in ./build/sqja3uK17MJxfC7AN8nGadBw9JK5BcrsNwNynsqP5Gih8M5Bm
Building spaces-cli in ./build/spaces-cli
Moving binary sqja3uK17MJxfC7AN8nGadBw9JK5BcrsNwNynsqP5Gih8M5Bm into plugin directory...
Cleaning up temporary files...
Adding virtual machine sqja3uK17MJxfC7AN8nGadBw9JK5BcrsNwNynsqP5Gih8M5Bm to installation registry...
Successfully installed luxfi/plugins-core:spacesvm@v0.0.4 in /Users/joshua.kim/go/src/github.com/luxfi/node/build/plugins/sqja3uK17MJxfC7AN8nGadBw9JK5BcrsNwNynsqP5Gih8M5Bm
Updating virtual machines...
Node at 127.0.0.1:9650/ext/admin was offline. Virtual machines will be available upon node startup.
Whitelisting subnet Ai42MkKqk8yjXFCpoHXw7rdTWSHiKEMqh5h8gbxwjgkCUfkrk...
Finished installing virtual machines for subnet Ai42MkKqk8yjXFCpoHXw7rdTWSHiKEMqh5h8gbxwjgkCUfkrk.
```

### Setting up Credentials for a Private Plugin Repository
You'll need to specify the `--credentials-file` flag which contains your github personal access token.

Example token file:
```
username: joshua-kim (for GitHub, this field doesn't matter. You can use your username as a placeholder)
password: <personal access token here>
```

Example command to download a subnet's VMs from a private repository:
```
lpm join-subnet --subnet=foobar --credentials-file=/home/joshua-kim/token
```
