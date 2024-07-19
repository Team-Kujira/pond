# pond

Pond is an easy way to set up a local Kujira development chain.

It creates one or more local Cosmos chains and connects them via [IBC](https://docs.cosmos.network/v0.45/ibc/overview.html), sets up price feeders for the on-chain oracle and deploys a set of Kujira core smart contracts.

Because Pond is meant to help builders on Kujira, its first and default chain will always be a Kujira chain and some of the commands (like `deploy` or `gov`) may only work here.

## Installation

Requires [docker](https://www.docker.com) and [golang](https://go.dev) `>=1.21`

```text
make install
```

## Init

Init creates all required configurations for validator nodes, price feeders and the IBC relayer.

The default configuration sets up a single Kujira chain with one validator node.

```text
pond init
```

### Nodes

Set the number of validator nodes on the main (Kujira) chain. The default value is one, maximum is nine.

```text
pond init --nodes 3
```

### Chains

If you need different or more partner chains, thats possible too. All chains will be connected to the first (Kujira) chain.

```text
pond init --chains cosmoshub
```

```text
pond init --chains kujira,kujira,terra2
```

### Contracts

Pond will deploy a set of Kujira core contracts on its first start. If you want addtional contracts, you can provide the list with:

```text
pond init --contracts kujira
```

Or deploy none:

```text
pond init --no-contracts
```

### Listen Address

Pond will forward all necessary ports to `localhost`. If you want to make your pond available from outside your machine and make it available for others, you can provide an ip address to bind to:

```text
pond init --listen 1.2.3.4
```

### Unbonding Time

The default unbonding time is set to 14 days. If you need to test staking related scenarios, you can set a custom time (in seconds).

Beware that this affects the IBC relayer and will freeze the clients if your Pond is stopped for a longer than the provided time.

```text
pond init --unbonding-time 300
```

### API/RPC URLs

Pond uses the [cosmos.directory](https://cosmos.directory/kujira) proxy to get data from public API/RPC nodes. Override them with:

```text
pond init --api-url https://my.api.node
```

```text
pond init --rpc-url https://my.rpc.node
```

### Local Binary

In case you need a custom Kujira version, you can use a local kujirad binary. The log output is written to `$HOME/.pond/kujira1-<N>/kujirad.log`

:warning: **Only works for kujirad >= v1.0.0**

```text
pond init --binary /path/to/my/kujirad
```

### Overrides

You can override default genesis parameters by providing a json file containing all the needed changes.

```text
pond init --overrides custom-settings.json
```

custom-settings.json:

```json
{
  "app_state": {
    "oracle": {
      "params": {
        "required_denoms": [
          "FOO",
          "BAR"
        ]
      }
    }
  }
}
```

## Start

Start your Pond

```text
pond start
```

## Stop

Stop your Pond

```text
pond stop
```

## Info

Retrieve infrastructure information

### Accounts

List all accounts and their addresses

```text
$ pond info accounts
chain  name     address
kujira deployer kujira1k3g54c2sc7g9mgzuzaukm9pvuzcjqy92nk9wse
kujira relayer  kujira1egssdw6et0pwdcdzl4nvyck68g3qcru99ynkjr
kujira test0    kujira1cyyzpxplxdzkeea7kwsydadg87357qnaww84dg
kujira test1    kujira18s5lynnmx37hq4wlrw9gdn68sg2uxp5r39mjh5
```

### Seed Phrase

List the seed phrase of a specific account

```text
$ pond info seed test0
notice oak worry limit wrap speak medal online prefer cluster roof addict wrist behave treat actual wasp year salad speed social layer crew genius
```

### Codes

List all deployed code ids and associated names

```text
$ pond info codes
id checksum          name
 1 418CF9A2…B2DFAA22 kujira_bow_xyk
 2 715934C4…49E9A7E9 kujira_bow_lsd
 3 F3B81230…9ABCC502 kujira_bow_stable
 4 98CC2EDA…CFF50C5F kujira_stable_mint
 5 8A6FA03E…BD42C198 kujira_fin
```

### Contracts

List all deployed contracts by code id

```text
$ pond info contracts
id address                                                           label
 1 kujira1e8rl4aawc44c2kqrx6urxktsf9h8k9sg9yufaxgegge2sn85vx7sdek0cz Bow ETH-USK
 1 kujira1narj7rjhmth6hzk8rmd3sqeus293tpj7r35n0dlm6lgle6zuevusl43a5j Bow KUJI-USK
 2 kujira1xfm4hyctm8zxjtjw7lsvtewnks36pekzt4x9fjjvhm8tr23vj42qp8m3au USK Controller
 3 kujira1gze5anmdc34plj9vaku3mn2cdurhd6r2680fr2xhvdcp32jzwl4q4t752w Fin ETH-USK
 3 kujira1nc8c8zktapz5y25jqfw4dlu8u0z0m2j5lhv4djrahvtr2dgekceqww683c Fin KUJI-USK
 3 kujira1ttcd5lk9xw2kenzxh7060h8ehyklqp5rx92j9p0ux36vsjhhrx9qmqr896 Fin USDC-USK
 3 kujira17gr3fgpwes8q8y0gt6rqjnqwe7p4dpel85nu57epedujdmuhug7sexs9fd Fin stETH-ETH
 4 kujira1pz4z5kakz60z738ghu4v8sc6qumuxcrezxafqhmzh77lw88y0vmqhu03uz Bow stETH-ETH
 5 kujira1yzd7hwez9yntpc3z34qcx5c7gduw65qk92a88u2pgpw0cxgdq7lsleea3r Bow USDC-USK
```

### URLs

List all chain and node specific URLs

```text
$ pond info urls
kujira1-1
 api    http://127.0.0.1:11117
 rpc    http://127.0.0.1:11157
 grpc   http://127.0.0.1:11190
 feeder http://127.0.0.1:11171/api/v1/prices
```

## Tx

Do Transaction on any of your chains. It will send the transaction to `kujira-1`if no `--chain-id` is provided.

```text
pond tx ibc-transfer transfer transfer channel-0 cosmos18s5lynnmx37hq4wlrw9gdn68sg2uxp5rqde267 123456789ukuji --from test0
```

## Query

Query any of your chains. It will query `kujira-1` by default if you don't provide a `--chain-id`.

```text
$ pond q bank balances cosmos18s5lynnmx37hq4wlrw9gdn68sg2uxp5rqde267 --chain-id cosmoshub-1
balances:
- amount: "123456789"
  denom: ibc/90F27756D300141BDF07B83E65401BDC58C05269B9BAE3ECB0B20FAB166BCF8F
- amount: "1000000000000"
  denom: uatom
pagination:
  next_key: null
  total: "0"
```

## Upgrade

Upgrade provides an easy way to test chain upgrades. For this to work, Pond creates an upgrade proposal, waits for the upgrade height and then restarts using the new binary.

:warning: **Only works in `--binary` mode**

## Government

Submit a gov proposal and optionally let all validators vote with the specified option.

```text
pond gov submit-proposal my-proposal.json
```

```text
pond gov submit-proposal my-proposal.json --vote yes
```

## Deploy

Deploy local wasm binaries or execute [plan files](##Planfiles) (a way to automate contract deployments)

```text
pond deploy myapp.wasm
```

```text
pond deploy myplan.json
```

## Code Registry

For the plan deployment to work, Pond stores the required wasm code information in the code registry and maps it to a human readable name which is needed in the plan files.

It ships with a default registry, containing most of the Kujira core apps and can be managed by using `pond registry` commands or editing `$HOME/.pond/registry.json` manually.

Every binary that is deployed via `pond deploy` will automatically added to the registry and referenced to its basename, so it can immediately be used for plan file deployments.

### Management

#### Update

Update the registry entry, in case the name or location of a code has changed

```text
pond registry update myapp.wasm --name myapp --source file:///tmp/myapp.wasm
```

#### Export / Import

In case you want to set up a new pond or apply your plan file on a different Pond, you can export your current registry and import it again

```text
pond export /tmp/myregistry.json
```

```text
pond import /tmp/myregistry.json
```

### Available sources

#### Mainnet

This downloads and deploys the code with the given code id from Kujira mainnet

```json
{
  "mycode": {
    "source": "kaiyo-1://12345"
  }
}
```

### Local Disk

This deploys locally stored code

```json
{
  "mycode": {
    "source": "file:///tmp/12345.wasm"
  }
}
```

## Planfiles

To make complex smart contract deployments easy to maintain, Pond offers the possibility to describe the order and parameters of your contract deployments and required denoms in json files and execute them. Pond then takes care of the creation of needed denoms and wasm code deployments as well as the instantiation of the contracts and needed contract executions. It keeps track all items and lets you access certain properties like contract address or denom path in every subsequent deployment via a simple template string (see "Planfile Syntax" for explanation).

Default plan files shipped with pond can be found in `$HOME/.pond/planfiles`.

```json
{
  "denom": "{{ .Denoms.USDC.Path }}",
  "address": "{{ .Contracts.my_contract.Address }}",
  "code_id": "{{ .CodeIds.my_contract }} | int"
}
```

Note: To be able to use the resulting code id as an integer value, you need to add `| int` after the template string. Otherwise the code id is provided as a string.

All deployments are done from the `deployer` account: `kujira1k3g54c2sc7g9mgzuzaukm9pvuzcjqy92nk9wse`

### Syntax

Plan files are simple json files that describe denoms and contract instantiations. They are designed to be as simple as possible and use the same instantiation message like the one that is used when executing it manually via `kujirad wasm execute`.

#### Denoms

Pond creates denoms that have a provided `nonce` as `factory/kujira1k3g54c2sc7g9mgzuzaukm9pvuzcjqy92nk9wse/{nonce}` from the `deployer` account. If `mint` is provided, it will also mint the specified amount of tokens into each `test*` account. All denoms with a provided `path` will be skipped.

Example:

```json
{
  "denoms": [
    {
      "name": "KUJI",
      "path": "ukuji"
    },
    {
      "name": "USDC",
      "nonce": "ibc/FE98AA...9576A9"
    },
    {
      "name": "POND",
      "nonce": "upond",
      "mint": "10_000_000"
    },
  ]
}
```

The above example will create the `POND` token as `factory/kujira1k3g54c2sc7g9mgzuzaukm9pvuzcjqy92nk9wse/upond` and mint 10 `POND` (10m upond) into each test wallet. `KUJI` and `USDC` have a path provided and therefore will be skipped.

It stores the path of all three denoms, which can be accessed in subsequent contract instantiations of that deployment via `{{ .Denoms.POND.Address }}` for example.

#### Codes

If you are working on contracts that change a lot during the development, updating the registry all the might become a tedious task. Therefore you can specify your sources directly in the plan file and Pond deploys them and updates the registry accordingly.

```json
{
  "codes":
    "my_project": "file:///path/to/my_project/artifacts/my_project-aarch64.wasm"
}
```

#### Contracts

Pond instantiates all contracts from the `deployer` account.

To speed up the deployments, Pond handles contracts in batches which combine all instantiations into a single transaction:

```json
{
  "contracts": [
    [
      {"name": "contract1", "msg": {}},
      {"name": "contract2", "msg": {}}
    ],
    [
      {"name": "contract3", "msg": {}}
    ]
  ]
}
```

The above example executes two batches. The first creating contract1 and contract2 in one transaction, the second creating contract3.

This way you speed up the deployment by maintaing a specific order of instantiations, if needed.

##### name

The name of your contract inside the plan file deployment run. This is needed to be able to refer to the contract in later instantiations.

##### code

The name of the wasm code stored in the registry. Pond will deploy the code, if it hasn't been yet.

##### label

A label for the newly created contract.

##### msg

The instantiation message needed to create the new contract.

##### funds

Some contracts need to be provided some amount of funds when they are created.

##### creates

Some contracts create tokens on their instantiation. To be able for Pond to make them usable in further steps, it needs to know their names.

##### actions

You can specify `/cosmwasm.wasm.v1.MsgExecuteContract` messages that are triggered after the contract is instantiated. This is useful, if you need to grant permissions for the newly created contract to a different contract for example.

### Example

The following example is a part of the `kujira` plan file, that is shipped and applied by default on the first start of each Pond instance (can be disabled with `--no-contracts`). It should showcase the use of the deployment order and templating.

```json
{
  "denoms": [
    {
      "name": "KUJI",
      "path": "ukuji"
    }
  ],
  "contracts": [
    [
      {
        "name": "kujira_stable_mint_usk",
        "code": "kujira_stable_mint",
        "label": "USK Controller",
        "funds": "10000000ukuji",
        "msg": {
          "denom": "uusk"
        },
        "creates": [
          {
            "name": "USK",
            "nonce": "uusk"
          }
        ]
      }
    ],
    [
      {
        "name": "kujira_fin_kuji_usk",
        "code": "kujira_fin",
        "label": "Fin KUJI-USK",
        "msg": {
          "denoms": [
            {
              "native": "{{ .Denoms.KUJI.Path }}"
            },
            {
              "native": "{{ .Denoms.USK.Path }}"
            }
          ],
          "price_precision": {
            "decimal_places": 4
          },
          "decimal_delta": 0,
          "fee_taker": "0.0015",
          "fee_maker": "0.00075"
        }
      }
    ],
    [
      {
        "name": "kujira_bow_kuji_usk",
        "code": "kujira_bow_xyk",
        "label": "Bow KUJI-USK",
        "funds": "10000000ukuji",
        "msg": {
          "fin_contract": "{{ .Contracts.kujira_fin_kuji_usk.Address }}",
          "intervals": [
            "0.01",
            "0.05"
          ],
          "fee": "0.1",
          "amp": "1"
        }
      }
    ]
  ]
}
```

In the above example plan file, Pond will at first check if it can find the referenced wasm codes in the code registry and deploy them if can't find them on-chain.

Next, it will create all tokens that have a provided `nonce` but no `path`. In this example, it crates no denom, because KUJI is provided with a `path`.

Now that all codes are deployed and needed denoms are created, it instantiates the "USK Controller". That will instantly create the USK token and therefore needs some initial `funds`. To be able to refer to the newly created USK address later, Pond retrieves and stores the path of all assets listed in the `creates` section.

It then instantiates a KUJI/USK Fin market and uses the retrieved path from the USK token, created before.

Last but not least it creates a Bow pool by referencing to the KUJI/USK Fin market via its name: `kujira_fin_kuji_usk`.
