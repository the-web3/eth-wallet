package config

import (
	"time"

	"github.com/urfave/cli/v2"

	"github.com/ethereum/go-ethereum/log"
	"github.com/the-web3/eth-wallet/flags"
)

const (
	defaultConfirmations    = 64
	defaultDepositInterval  = 5000
	defaultWithdrawInterval = 500
	defaultCollectInterval  = 500
	defaultColdInterval     = 500
	defaultBlocksStep       = 500
)

type Config struct {
	Migrations     string
	Chain          ChainConfig
	MasterDB       DBConfig
	SlaveDB        DBConfig
	SlaveDbEnable  bool
	ApiCacheEnable bool
	CacheConfig    CacheConfig
	RpcServer      ServerConfig
	HTTPServer     ServerConfig
	MetricsServer  ServerConfig
}

type ChainConfig struct {
	ChainID          uint
	RpcUrl           string
	StartingHeight   uint
	Confirmations    uint
	DepositInterval  uint
	WithdrawInterval uint
	CollectInterval  uint
	ColdInterval     uint
	BlocksStep       uint
}

type DBConfig struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
}

type CacheConfig struct {
	ListSize         int
	DetailSize       int
	ListExpireTime   time.Duration
	DetailExpireTime time.Duration
}

type ServerConfig struct {
	Host string
	Port int
}

func LoadConfig(cliCtx *cli.Context) (Config, error) {
	var cfg Config
	cfg = NewConfig(cliCtx)

	if cfg.Chain.Confirmations == 0 {
		cfg.Chain.Confirmations = defaultConfirmations
	}

	if cfg.Chain.DepositInterval == 0 {
		cfg.Chain.DepositInterval = defaultDepositInterval
	}

	if cfg.Chain.WithdrawInterval == 0 {
		cfg.Chain.WithdrawInterval = defaultWithdrawInterval
	}

	if cfg.Chain.CollectInterval == 0 {
		cfg.Chain.CollectInterval = defaultCollectInterval
	}

	if cfg.Chain.ColdInterval == 0 {
		cfg.Chain.ColdInterval = defaultColdInterval
	}

	if cfg.Chain.BlocksStep == 0 {
		cfg.Chain.BlocksStep = defaultBlocksStep
	}

	log.Info("loaded chain config", "config", cfg.Chain)
	return cfg, nil
}

func NewConfig(ctx *cli.Context) Config {
	return Config{
		Migrations: ctx.String(flags.MigrationsFlag.Name),
		Chain: ChainConfig{
			ChainID:          ctx.Uint(flags.ChainIdFlag.Name),
			RpcUrl:           ctx.String(flags.RpcUrlFlag.Name),
			StartingHeight:   ctx.Uint(flags.StartingHeightFlag.Name),
			Confirmations:    ctx.Uint(flags.ConfirmationsFlag.Name),
			DepositInterval:  ctx.Uint(flags.DepositIntervalFlag.Name),
			WithdrawInterval: ctx.Uint(flags.WithdrawIntervalFlag.Name),
			CollectInterval:  ctx.Uint(flags.CollectIntervalFlag.Name),
			ColdInterval:     ctx.Uint(flags.ColdIntervalFlag.Name),
			BlocksStep:       ctx.Uint(flags.BlocksStepFlag.Name),
		},
		MasterDB: DBConfig{
			Host:     ctx.String(flags.MasterDbHostFlag.Name),
			Port:     ctx.Int(flags.MasterDbPortFlag.Name),
			Name:     ctx.String(flags.MasterDbNameFlag.Name),
			User:     ctx.String(flags.MasterDbUserFlag.Name),
			Password: ctx.String(flags.MasterDbPasswordFlag.Name),
		},
		SlaveDB: DBConfig{
			Host:     ctx.String(flags.SlaveDbHostFlag.Name),
			Port:     ctx.Int(flags.SlaveDbPortFlag.Name),
			Name:     ctx.String(flags.SlaveDbNameFlag.Name),
			User:     ctx.String(flags.SlaveDbUserFlag.Name),
			Password: ctx.String(flags.SlaveDbPasswordFlag.Name),
		},
		SlaveDbEnable:  ctx.Bool(flags.SlaveDbEnableFlag.Name),
		ApiCacheEnable: ctx.Bool(flags.ApiCacheEnableFlag.Name),
		CacheConfig: CacheConfig{
			ListSize:         ctx.Int(flags.ApiCacheListSizeFlag.Name),
			DetailSize:       ctx.Int(flags.ApiCacheDetailSizeFlag.Name),
			ListExpireTime:   ctx.Duration(flags.ApiCacheListExpireTimeFlag.Name),
			DetailExpireTime: ctx.Duration(flags.ApiCacheDetailExpireTimeFlag.Name),
		},
		RpcServer: ServerConfig{
			Host: ctx.String(flags.RpcHostFlag.Name),
			Port: ctx.Int(flags.RpcPortFlag.Name),
		},
		HTTPServer: ServerConfig{
			Host: ctx.String(flags.HttpHostFlag.Name),
			Port: ctx.Int(flags.HttpPortFlag.Name),
		},
		MetricsServer: ServerConfig{
			Host: ctx.String(flags.MetricsHostFlag.Name),
			Port: ctx.Int(flags.MetricsPortFlag.Name),
		},
	}
}
