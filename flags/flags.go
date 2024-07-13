package flags

import (
	"github.com/urfave/cli/v2"
	"time"
)

const evnVarPrefix = "ETH_WALLET"

func prefixEnvVars(name string) []string {
	return []string{evnVarPrefix + "_" + name}
}

var (
	MigrationsFlag = &cli.StringFlag{
		Name:    "migrations-dir",
		Value:   "./migrations",
		Usage:   "path for database migrations",
		EnvVars: prefixEnvVars("MIGRATIONS_DIR"),
	}
	// chain flags
	ChainIdFlag = &cli.UintFlag{
		Name:     "chain-id",
		Usage:    "The port of the api",
		EnvVars:  prefixEnvVars("CHAIN_ID"),
		Value:    1,
		Required: true,
	}
	RpcUrlFlag = &cli.StringFlag{
		Name:     "rpc-url",
		Usage:    "HTTP provider URL for chain",
		EnvVars:  prefixEnvVars("RPC_RUL"),
		Required: true,
	}
	StartingHeightFlag = &cli.UintFlag{
		Name:    "starting-height",
		Usage:   "The starting height of chain",
		EnvVars: prefixEnvVars("STARTING_HEIGHT"),
		Value:   0,
	}
	ConfirmationsFlag = &cli.UintFlag{
		Name:    "confirmations",
		Usage:   "The confirmation depth of l1",
		EnvVars: prefixEnvVars("CONFIRMATIONS"),
		Value:   64,
	}
	DepositIntervalFlag = &cli.DurationFlag{
		Name:    "deposit-interval",
		Usage:   "The interval of l1 synchronization",
		EnvVars: prefixEnvVars("DEPOSIT_INTERVAL"),
		Value:   time.Second * 5,
	}
	WithdrawIntervalFlag = &cli.DurationFlag{
		Name:    "withdraw-interval",
		Usage:   "The interval of token withdraw",
		EnvVars: prefixEnvVars("WITHDRAW_INTERVAL"),
		Value:   time.Second * 5,
	}
	CollectIntervalFlag = &cli.DurationFlag{
		Name:    "collect-interval",
		Usage:   "The interval of collect wallet funding",
		EnvVars: prefixEnvVars("COLLECT_INTERVAL"),
		Value:   time.Second * 5,
	}
	ColdIntervalFlag = &cli.DurationFlag{
		Name:    "cold-interval",
		Usage:   "The interval of transfer token to cold wallet",
		EnvVars: prefixEnvVars("COLD_INTERVAL"),
		Value:   time.Second * 5,
	}
	BlocksStepFlag = &cli.UintFlag{
		Name:    "blocks-step",
		Usage:   "Scanner blocks step",
		EnvVars: prefixEnvVars("BLOCKS_STEP"),
		Value:   500,
	}
	// Rest api flags
	HttpHostFlag = &cli.StringFlag{
		Name:     "http-host",
		Usage:    "The host of the api",
		EnvVars:  prefixEnvVars("HTTP_HOST"),
		Required: true,
	}
	HttpPortFlag = &cli.IntFlag{
		Name:     "http-port",
		Usage:    "The port of the api",
		EnvVars:  prefixEnvVars("HTTP_PORT"),
		Value:    8987,
		Required: true,
	}

	// rpc api flags
	RpcHostFlag = &cli.StringFlag{
		Name:     "rpc-host",
		Usage:    "The host of the rpc",
		EnvVars:  prefixEnvVars("RPC_HOST"),
		Required: true,
	}
	RpcPortFlag = &cli.IntFlag{
		Name:     "rpc-port",
		Usage:    "The port of the rpc",
		EnvVars:  prefixEnvVars("RPC_PORT"),
		Value:    8987,
		Required: true,
	}
	// Metrics flags
	MetricsHostFlag = &cli.StringFlag{
		Name:     "metrics-host",
		Usage:    "The host of the metrics",
		EnvVars:  prefixEnvVars("METRICS_HOST"),
		Required: true,
	}
	MetricsPortFlag = &cli.IntFlag{
		Name:     "metrics-port",
		Usage:    "The port of the metrics",
		EnvVars:  prefixEnvVars("METRICS_PORT"),
		Value:    7214,
		Required: true,
	}

	SlaveDbEnableFlag = &cli.BoolFlag{
		Name:     "slave-db-enable",
		Usage:    "Whether to use slave db",
		EnvVars:  prefixEnvVars("SLAVE_DB_ENABLE"),
		Required: true,
	}
	ApiCacheEnableFlag = &cli.BoolFlag{
		Name:     "api-cache-enable",
		Usage:    "api cache enable",
		EnvVars:  prefixEnvVars("API_CACHE_ENABLE"),
		Required: true,
	}

	// MasterDb Flags
	MasterDbHostFlag = &cli.StringFlag{
		Name:     "master-db-host",
		Usage:    "The host of the master database",
		EnvVars:  prefixEnvVars("MASTER_DB_HOST"),
		Required: true,
	}
	MasterDbPortFlag = &cli.IntFlag{
		Name:     "master-db-port",
		Usage:    "The port of the master database",
		EnvVars:  prefixEnvVars("MASTER_DB_PORT"),
		Required: true,
	}
	MasterDbUserFlag = &cli.StringFlag{
		Name:     "master-db-user",
		Usage:    "The user of the master database",
		EnvVars:  prefixEnvVars("MASTER_DB_USER"),
		Required: true,
	}
	MasterDbPasswordFlag = &cli.StringFlag{
		Name:     "master-db-password",
		Usage:    "The host of the master database",
		EnvVars:  prefixEnvVars("MASTER_DB_PASSWORD"),
		Required: true,
	}
	MasterDbNameFlag = &cli.StringFlag{
		Name:     "master-db-name",
		Usage:    "The db name of the master database",
		EnvVars:  prefixEnvVars("MASTER_DB_NAME"),
		Required: true,
	}

	// Slave DB  flags
	SlaveDbHostFlag = &cli.StringFlag{
		Name:    "slave-db-host",
		Usage:   "The host of the slave database",
		EnvVars: prefixEnvVars("SLAVE_DB_HOST"),
	}
	SlaveDbPortFlag = &cli.IntFlag{
		Name:    "slave-db-port",
		Usage:   "The port of the slave database",
		EnvVars: prefixEnvVars("SLAVE_DB_PORT"),
	}
	SlaveDbUserFlag = &cli.StringFlag{
		Name:    "slave-db-user",
		Usage:   "The user of the slave database",
		EnvVars: prefixEnvVars("SLAVE_DB_USER"),
	}
	SlaveDbPasswordFlag = &cli.StringFlag{
		Name:    "slave-db-password",
		Usage:   "The host of the slave database",
		EnvVars: prefixEnvVars("SLAVE_DB_PASSWORD"),
	}
	SlaveDbNameFlag = &cli.StringFlag{
		Name:    "slave-db-name",
		Usage:   "The db name of the slave database",
		EnvVars: prefixEnvVars("SLAVE_DB_NAME"),
	}

	// cache flags
	ApiCacheListSizeFlag = &cli.UintFlag{
		Name:    "api-cache-list-size",
		Usage:   "The size of the api cache list",
		EnvVars: prefixEnvVars("API_CACHE_LIST_SIZE"),
	}
	ApiCacheDetailSizeFlag = &cli.UintFlag{
		Name:    "api-cache-detail-size",
		Usage:   "The size of the api cache detail",
		EnvVars: prefixEnvVars("API_CACHE_LIST_DETAIL"),
	}
	ApiCacheListExpireTimeFlag = &cli.DurationFlag{
		Name:    "api-cache-list-expire-time",
		Usage:   "The interval of collect wallet funding",
		EnvVars: prefixEnvVars("API_CACHE_LIST_EXPIRE_TIME"),
		Value:   time.Minute * 30,
	}
	ApiCacheDetailExpireTimeFlag = &cli.DurationFlag{
		Name:    "api-cache-detail-expire-time",
		Usage:   "The interval of collect wallet funding",
		EnvVars: prefixEnvVars("API_CACHE_DETAIL_EXPIRE_TIME"),
		Value:   time.Minute * 30,
	}
)

var requireFlags = []cli.Flag{
	MigrationsFlag,
	ChainIdFlag,
	RpcUrlFlag,
	StartingHeightFlag,
	ConfirmationsFlag,
	DepositIntervalFlag,
	WithdrawIntervalFlag,
	CollectIntervalFlag,
	ColdIntervalFlag,
	BlocksStepFlag,
	HttpHostFlag,
	HttpPortFlag,
	RpcHostFlag,
	RpcPortFlag,
	MetricsPortFlag,
	MetricsHostFlag,
	SlaveDbEnableFlag,
	MasterDbHostFlag,
	MasterDbPortFlag,
	MasterDbUserFlag,
	MasterDbPasswordFlag,
	MasterDbNameFlag,
}

var optionalFlags = []cli.Flag{
	SlaveDbHostFlag,
	SlaveDbPortFlag,
	SlaveDbUserFlag,
	SlaveDbPasswordFlag,
	SlaveDbNameFlag,
	ApiCacheListSizeFlag,
	ApiCacheDetailSizeFlag,
	ApiCacheListExpireTimeFlag,
	ApiCacheDetailExpireTimeFlag,
}

func init() {
	Flags = append(requireFlags, optionalFlags...)
}

var Flags []cli.Flag
