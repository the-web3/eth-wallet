<!--
parent:
  order: false
-->

<div align="center">
  <h1> Ethereum Exchange Wallet </h1>
</div>

<div align="center">
  <a href="https://github.com/the-web3/eth-wallet/releases/latest">
    <img alt="Version" src="https://img.shields.io/github/tag/the-web3/eth-wallet.svg" />
  </a>
  <a href="https://github.com/the-web3/eth-wallet/blob/main/LICENSE">
    <img alt="License: Apache-2.0" src="https://img.shields.io/github/license/the-web3/eth-wallet.svg" />
  </a>
  <a href="https://pkg.go.dev/github.com/the-web3/eth-wallet">
    <img alt="GoDoc" src="https://godoc.org/github.com/the-web3/eth-wallet?status.svg" />
  </a>
  <a href="https://goreportcard.com/report/github.com/the-web3/eth-wallet">
    <img alt="Go report card" src="https://goreportcard.com/badge/github.com/the-web3/eth-wallet"/>
  </a>
</div>

Ethereum centralized wallet include deposit, withdraw, collection and tranfer token to cold wallet

**Note**: Requires [Go 1.21+](https://golang.org/dl/)

## Installation

For prerequisites and detailed build instructions please read the [Installation](https://github.com/the-web3/eth-wallet/) instructions. Once the dependencies are installed, run:

```bash
make eth-wallet
```

Or check out the latest [release](https://github.com/the-web3/eth-wallet).

## Config env
```
ETH_WALLET_MIGRATIONS_DIR="./migrations"

ETH_WALLET_CHAIN_ID=17000
ETH_WALLET_RPC_RUL="please type your ethereum rpc url"
ETH_WALLET_STARTING_HEIGHT=1960574
ETH_WALLET_CONFIRMATIONS=32
ETH_WALLET_DEPOSIT_INTERVAL=5s
ETH_WALLET_WITHDRAW_INTERVAL=5s
ETH_WALLET_COLLECT_INTERVAL=5s
ETH_WALLET_BLOCKS_STEP=5

ETH_WALLET_HTTP_PORT=8989
ETH_WALLET_HTTP_HOST="127.0.0.1"
ETH_WALLET_RPC_PORT=8980
ETH_WALLET_RPC_HOST="127.0.0.1"
ETH_WALLET_METRICS_PORT=8990
ETH_WALLET_METRICS_HOST="127.0.0.1"

ETH_WALLET_SLAVE_DB_ENABLE=false

ETH_WALLET_MASTER_DB_HOST="127.0.0.1"
ETH_WALLET_MASTER_DB_PORT=5432
ETH_WALLET_MASTER_DB_USER="guoshijiang"
ETH_WALLET_MASTER_DB_PASSWORD=""
ETH_WALLET_MASTER_DB_NAME="eth_wallet"

ETH_WALLET_SLAVE_DB_HOST="127.0.0.1"
ETH_WALLET_SLAVE_DB_PORT=5432
ETH_WALLET_SLAVE_DB_USER="guoshijiang"
ETH_WALLET_SLAVE_DB_PASSWORD=""
ETH_WALLET_SLAVE_DB_NAME="eth_wallet"

ETH_WALLET_API_CACHE_LIST_SIZE=0
ETH_WALLET_API_CACHE_LIST_DETAIL=0
ETH_WALLET_API_CACHE_LIST_EXPIRE_TIME=0
ETH_WALLET_API_CACHE_DETAIL_EXPIRE_TIME=0
```

## Quick Start

### 1.create database 
```
create database eth_wallet;
```

### 2.migrate your database

#### command
```
./eth-wallet migrate
```
execute result

#### check

```
postgres=# \c eth_wallet
您现在已经连接到数据库 "eth_wallet",用户 "guoshijiang".
eth_wallet=# \d
                    关联列表
 架构模式 |     名称     |  类型  |   拥有者
----------+--------------+--------+-------------
 public   | addresses    | 数据表 | guoshijiang
 public   | balances     | 数据表 | guoshijiang
 public   | blocks       | 数据表 | guoshijiang
 public   | deposits     | 数据表 | guoshijiang
 public   | tokens       | 数据表 | guoshijiang
 public   | transactions | 数据表 | guoshijiang
 public   | withdraws    | 数据表 | guoshijiang
(7 行记录)

eth_wallet=#
```

### 3.batch address generate

#### command
```
./eth-wallet generate-address
```

#### check

```
eth_wallet=# select count(*) from addresses;
 count
-------
   100
(1 行记录)

eth_wallet=#
```

### startup wallet


#### command
```
./eth-wallet wallet
```

#### check

```
INFO [07-20|16:26:52.255] exec wallet sync
INFO [07-20|16:26:52.255] loaded chain config                      config="{ChainID:17000 RpcUrl:https://eth-holesky.g.alchemy.com/v2/BvSZ5ZfdIwB-5SDXMz8PfGcbICYQqwrl StartingHeight:1964509 Confirmations:64 DepositInterval:5000 WithdrawInterval:500 CollectInterval:500 ColdInterval:500 BlocksStep:5}"

2024/07/20 16:26:52 /Users/guoshijiang/theweb3/eth-wallet/database/blocks.go:56 record not found
[5.827ms] [rows:0] SELECT * FROM "blocks" ORDER BY number DESC LIMIT 1
INFO [07-20|16:26:52.758] no sync indexed state starting from supplied ethereum height height=1,964,509
INFO [07-20|16:26:53.471] start deposit......
INFO [07-20|16:26:53.472] start withdraw......
```




