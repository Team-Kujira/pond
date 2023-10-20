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
