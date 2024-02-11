# pond

Pond is an easy way to set up a local Kujira development chain. It uses docker containers to set up two local Kujira chains, price feeder and an IBC relayer connecting both chains.

The second chain is meant to test IBC related things and therefore has only one validator and no price feeder.

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

## Accounts

### Pre-funded test wallets

#### kujira1cyyzpxplxdzkeea7kwsydadg87357qnaww84dg

```text
notice oak worry limit wrap speak medal online prefer cluster roof addict wrist behave treat actual wasp year salad speed social layer crew genius
```

#### kujira18s5lynnmx37hq4wlrw9gdn68sg2uxp5r39mjh5

```text
quality vacuum heart guard buzz spike sight swarm shove special gym robust assume sudden deposit grid alcohol choice devote leader tilt noodle tide penalty
```

#### kujira1qwexv7c6sm95lwhzn9027vyu2ccneaqa5xl0d9

```text
symbol force gallery make bulk round subway violin worry mixture penalty kingdom boring survey tool fringe patrol sausage hard admit remember broken alien absorb
```

#### kujira14hcxlnwlqtq75ttaxf674vk6mafspg8xhmzm0f

```text
bounce success option birth apple portion aunt rural episode solution hockey pencil lend session cause hedgehog slender journey system canvas decorate razor catch empty
```

#### kujira12rr534cer5c0vj53eq4y32lcwguyy7nn5c753n

```text
second render cat sing soup reward cluster island bench diet lumber grocery repeat balcony perfect diesel stumble piano distance caught occur example ozone loyal
```

#### kujira1nt33cjd5auzh36syym6azgc8tve0jlvkxq3kfc

```text
spatial forest elevator battle also spoon fun skirt flight initial nasty transfer glory palm drama gossip remove fan joke shove label dune debate quick
```

#### kujira10qfrpash5g2vk3hppvu45x0g860czur8s69vah

```text
noble width taxi input there patrol clown public spell aunt wish punch moment will misery eight excess arena pen turtle minimum grain vague inmate
```

#### kujira1f4tvsdukfwh6s9swrc24gkuz23tp8pd3qkjuj9

```text
cream sport mango believe inhale text fish rely elegant below earth april wall rug ritual blossom cherry detail length blind digital proof identify ride
```

#### kujira1myv43sqgnj5sm4zl98ftl45af9cfzk7nwph6m0

```text
index light average senior silent limit usual local involve delay update rack cause inmate wall render magnet common feature laundry exact casual resource hundred
```

#### kujira14gs9zqh8m49yy9kscjqu9h72exyf295asmt7nw

```text
prefer forget visit mistake mixture feel eyebrow autumn shop pair address airport diesel street pass vague innocent poem method awful require hurry unhappy shoulder
```
