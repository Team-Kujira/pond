# pond

Pond is an easy way to set up a local Kujira development chain.

## Installation

```bash
sudo cp pond /usr/local/bin/
```

## Usage

### Init new pond

The init step creates the validator and price feeder config needed to run a local Kujira chain and stores it in `$HOME/.pond`.

```bash
pond init --nodes 1
```

### Start pond

```bash
pond start
```

### Stop pond

```bash
pond stop
```

### Show pond information

```bash
pond info
```

### Dev Containers

To open the repository in a VS Code Devcontainer:
- Use VS Code on your local machine
- Install the [Dev Containers VS Code extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers)
- Click on the bottom left and choose to Reopen in Container

[!WARNING]
Do not use it in Github Codespaces as the Kujira apps will not see it at the required IP Addresses

This will install:
1. Pond
2. Rust with cosmwasm and wasmd
3. Node
4. Go