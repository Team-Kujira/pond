# First steps

This tutorial is meant to give you some examples of how to interact with Pond and some of Kujiras core dapps.

## Init & start Pond

```text
pond init && pond start
```

## Mint USK

The standard Pond deployment comes with the possibility to mint USK for ETH. The ETH in this case is just locally created factory asset, but it uses the real ETH oracle price. Each test account should have 10 ETH in its wallet (1 ETH = 10**18 aeth)

```text
$ pond q bank balances kujira1cyyzpxplxdzkeea7kwsydadg87357qnaww84dg
balances:
- amount: "10000000000000000000"
  denom: factory/kujira1k3g54c2sc7g9mgzuzaukm9pvuzcjqy92nk9wse/aeth
- amount: "10000000000000000000"
  denom: factory/kujira1k3g54c2sc7g9mgzuzaukm9pvuzcjqy92nk9wse/asteth
- amount: "100000000000"
  denom: factory/kujira1k3g54c2sc7g9mgzuzaukm9pvuzcjqy92nk9wse/uusdc
- amount: "1000000000000"
  denom: ukuji
pagination:
  total: "4"
```

### Deposit

To be able to use that ETH as collateral, we need to deposit some of it into the USK contract `USK Market ETH`. You can find the address with `pond info contracts`.

```text
pond tx wasm execute kujira1lcqrewe4fwcs7fx3y2uwxx6afh48833zlcaxjn7tpdragzcmhuhs083qj5 '{"deposit":{"address":"kujira1cyyzpxplxdzkeea7kwsydadg87357qnaww84dg"}}' --from test0 --amount 5000000000000000000factory/kujira1k3g54c2sc7g9mgzuzaukm9pvuzcjqy92nk9wse/aeth
```

We can query the contract and verify our deposit:

```text
pond q wasm contract-state smart kujira1lcqrewe4fwcs7fx3y2uwxx6afh48833zlcaxjn7tpdragzcmhuhs083qj5 '{"position":{"address":"kujira1cyyzpxplxdzkeea7kwsydadg87357qnaww84dg"}}'
data:
  deposit_amount: "5000000000000000000"
  interest_amount: "0"
  liquidation_price: null
  mint_amount: "0"
  owner: kujira1cyyzpxplxdzkeea7kwsydadg87357qnaww84dg
```

### Mint

Now that we have verified that the ETH has arrived in the contract, we can mint (borrow) USK for it:

```text
pond tx wasm execute kujira1lcqrewe4fwcs7fx3y2uwxx6afh48833zlcaxjn7tpdragzcmhuhs083qj5 '{"mint":{"amount":"5000000000","recipient":"kujira1cyyzpxplxdzkeea7kwsydadg87357qnaww84dg"}}' --from test0 --gas auto --gas-adjustment 2
```

We now just have received 4995 USK in our wallet:

```test
pond q bank balances kujira1cyyzpxplxdzkeea7kwsydadg87357qnaww84dg        
balances:
- amount: "5000000000000000000"
  denom: factory/kujira1k3g54c2sc7g9mgzuzaukm9pvuzcjqy92nk9wse/aeth
- amount: "10000000000000000000"
  denom: factory/kujira1k3g54c2sc7g9mgzuzaukm9pvuzcjqy92nk9wse/asteth
- amount: "100000000000"
  denom: factory/kujira1k3g54c2sc7g9mgzuzaukm9pvuzcjqy92nk9wse/uusdc
- amount: "4995000000"
  denom: factory/kujira1xfm4hyctm8zxjtjw7lsvtewnks36pekzt4x9fjjvhm8tr23vj42qp8m3au/uusk
- amount: "1000000000000"
  denom: ukuji
pagination:
  total: "5"
```

## Provide liquidity

Now with 4995 USK in our wallet, why not providing some liquidity for the KUJI-USK pool on Bow?

```text
pond tx wasm execute kujira1narj7rjhmth6hzk8rmd3sqeus293tpj7r35n0dlm6lgle6zuevusl43a5j '{"deposit":{}}' --from test0 --amount 2000000000factory/kujira1xfm4hyctm8zxjtjw7lsvtewnks36pekzt4x9fjjvhm8tr23vj42qp8m3au/uusk,1000000000ukuji --gas auto --gas-adjustment 2
```

To receive fees from our LP position, we also need to stake the LP tokens we just received (`1414213562factory/kujira1narj7rjhmth6hzk8rmd3sqeus293tpj7r35n0dlm6lgle6zuevusl43a5j/ulp`)

```text
pond tx wasm execute kujira1wun86nqrwl5cnggsf37pujztlr824c6u3ssnr0akduz74u4ctp7s8x7hf4 '{"stake":{"addr":"kujira1cyyzpxplxdzkeea7kwsydadg87357qnaww84dg"}}' --from test0 --amount 1414213562factory/kujira1narj7rjhmth6hzk8rmd3sqeus293tpj7r35n0dlm6lgle6zuevusl43a5j/ulp --gas auto --gas-adjustment 2
```

```text
pond q wasm contract-state smart kujira1wun86nqrwl5cnggsf37pujztlr824c6u3ssnr0akduz74u4ctp7s8x7hf4 '{"fills":{"denom":"factory/kujira1narj7rjhmth6hzk8rmd3sqeus293tpj7r35n0dlm6lgle6zuevusl43a5j/ulp","addr":"kujira1cyyzpxplxdzkeea7kwsydadg87357qnaww84dg"}}'
```

## Lend

Lets provide another part of our USK for lending on Ghost

```text
pond tx wasm execute kujira1fph5mhsdq9lrayvqetwfsqmk6vz7ac5mrrcl5a9ry5n4strv3hkq6vy6xd '{"deposit":{}}' --from test0 --amount 2000000000factory/kujira1xfm4hyctm8zxjtjw7lsvtewnks36pekzt4x9fjjvhm8tr23vj42qp8m3au/uusk --gas auto --gas-adjustment 2
```

## Borrow

Now, lets assume another person wants to borrow some of our provided USK on Ghost. Therefore we need to deposit some collateral to borrow against:

```text
pond tx wasm execute kujira1w2fz9uh4478ms6vd7gtl8jwxneejs52xvfj4alwrh75x37rp9res0yrerm '{"deposit":{}}' --from test1 --amount 1000000000ukuji --gas auto --gas-adjustment 2
```

And then borrow some USK:

```text
pond tx wasm execute kujira1w2fz9uh4478ms6vd7gtl8jwxneejs52xvfj4alwrh75x37rp9res0yrerm '{"borrow":{"amount":"500000000"}}' --from test1 --gas auto --gas-adjustment 2
```

We now have successfully borrowed 499 USK:

```text
pond q bank balances kujira18s5lynnmx37hq4wlrw9gdn68sg2uxp5r39mjh5        

balances:
- amount: "10000000000000000000"
  denom: factory/kujira1k3g54c2sc7g9mgzuzaukm9pvuzcjqy92nk9wse/aeth
- amount: "10000000000000000000"
  denom: factory/kujira1k3g54c2sc7g9mgzuzaukm9pvuzcjqy92nk9wse/asteth
- amount: "100000000000"
  denom: factory/kujira1k3g54c2sc7g9mgzuzaukm9pvuzcjqy92nk9wse/uusdc
- amount: "499000000"
  denom: factory/kujira1xfm4hyctm8zxjtjw7lsvtewnks36pekzt4x9fjjvhm8tr23vj42qp8m3au/uusk
- amount: "999000000000"
  denom: ukuji
pagination:
  total: "5"
```

Lets just buy some more KUJI with it at market price:

```text
pond tx wasm execute kujira1nc8c8zktapz5y25jqfw4dlu8u0z0m2j5lhv4djrahvtr2dgekceqww683c '{"swap":{}}' --from test1 --amount "499000000factory/kujira1xfm4hyctm8zxjtjw7lsvtewnks36pekzt4x9fjjvhm8tr23vj42qp8m3au/uusk" --gas auto --gas-adjustment 2
```

Which results in about 181.2 more KUJI in our wallet:

```text
pond q bank balances kujira18s5lynnmx37hq4wlrw9gdn68sg2uxp5r39mjh5        

balances:
- amount: "10000000000000000000"
  denom: factory/kujira1k3g54c2sc7g9mgzuzaukm9pvuzcjqy92nk9wse/aeth
- amount: "10000000000000000000"
  denom: factory/kujira1k3g54c2sc7g9mgzuzaukm9pvuzcjqy92nk9wse/asteth
- amount: "100000000000"
  denom: factory/kujira1k3g54c2sc7g9mgzuzaukm9pvuzcjqy92nk9wse/uusdc
- amount: "999181221923"
  denom: ukuji
pagination:
  total: "4"
```
