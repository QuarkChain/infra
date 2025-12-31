package app

import (
	"errors"
	"math"

	"github.com/urfave/cli/v2"

	opservice "github.com/ethereum-optimism/optimism/op-service"
	oplog "github.com/ethereum-optimism/optimism/op-service/log"
	opmetrics "github.com/ethereum-optimism/optimism/op-service/metrics"
	"github.com/ethereum-optimism/optimism/op-service/oppprof"
	oprpc "github.com/ethereum-optimism/optimism/op-service/rpc"
	optls "github.com/ethereum-optimism/optimism/op-service/tls"
)

const (
	ServiceConfigPathFlagName = "config"
	ClientEndpointFlagName    = "endpoint"
)

type HealthzConfig struct {
	Enabled    bool
	ListenAddr string
	ListenPort int
}

func (c HealthzConfig) Check() error {
	if c.ListenPort < 0 || c.ListenPort > math.MaxUint16 {
		return errors.New("invalid healthz port")
	}

	return nil
}

const (
	HealthzEnabledFlagName    = "healthz.enabled"
	HealthzListenAddrFlagName = "healthz.addr"
	HealthzListenPortFlagName = "healthz.port"
)

func HealthzCLIFlags(envPrefix string) []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:    HealthzEnabledFlagName,
			Usage:   "Enable the health check server",
			EnvVars: opservice.PrefixEnvVar(envPrefix, "HEALTHZ_ENABLED"),
		},
		&cli.StringFlag{
			Name:    HealthzListenAddrFlagName,
			Usage:   "Health check server listening address",
			Value:   "0.0.0.0",
			EnvVars: opservice.PrefixEnvVar(envPrefix, "HEALTHZ_ADDR"),
		},
		&cli.IntFlag{
			Name:    HealthzListenPortFlagName,
			Usage:   "Health check server listening port",
			Value:   7600,
			EnvVars: opservice.PrefixEnvVar(envPrefix, "HEALTHZ_PORT"),
		},
	}
}

func ReadHealthzCLIConfig(ctx *cli.Context) HealthzConfig {
	return HealthzConfig{
		Enabled:    ctx.Bool(HealthzEnabledFlagName),
		ListenAddr: ctx.String(HealthzListenAddrFlagName),
		ListenPort: ctx.Int(HealthzListenPortFlagName),
	}
}

func CLIFlags(envPrefix string) []cli.Flag {
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:    ServiceConfigPathFlagName,
			Usage:   "Signer service configuration file path",
			Value:   "config.yaml",
			EnvVars: opservice.PrefixEnvVar(envPrefix, "SERVICE_CONFIG"),
		},
	}
	flags = append(flags, oprpc.CLIFlags(envPrefix)...)
	flags = append(flags, oplog.CLIFlags(envPrefix)...)
	flags = append(flags, opmetrics.CLIFlags(envPrefix)...)
	flags = append(flags, oppprof.CLIFlags(envPrefix)...)
	flags = append(flags, optls.CLIFlags(envPrefix)...)
	flags = append(flags, HealthzCLIFlags(envPrefix)...)
	return flags
}

func ClientSignCLIFlags(envPrefix string) []cli.Flag {
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:    ClientEndpointFlagName,
			Usage:   "Signer endpoint the client will connect to",
			Value:   "http://localhost:8080",
			EnvVars: opservice.PrefixEnvVar(envPrefix, "CLIENT_ENDPOINT"),
		},
	}
	return flags
}

type Config struct {
	ClientEndpoint    string
	ServiceConfigPath string

	TLSConfig     optls.CLIConfig
	RPCConfig     oprpc.CLIConfig
	LogConfig     oplog.CLIConfig
	MetricsConfig opmetrics.CLIConfig
	PprofConfig   oppprof.CLIConfig
	HealthzConfig HealthzConfig
}

func (c Config) Check() error {
	if err := c.RPCConfig.Check(); err != nil {
		return err
	}
	if err := c.MetricsConfig.Check(); err != nil {
		return err
	}
	if err := c.PprofConfig.Check(); err != nil {
		return err
	}
	if err := c.TLSConfig.Check(); err != nil {
		return err
	}
	if err := c.HealthzConfig.Check(); err != nil {
		return err
	}
	return nil
}

func NewConfig(ctx *cli.Context) *Config {
	return &Config{
		ClientEndpoint:    ctx.String(ClientEndpointFlagName),
		ServiceConfigPath: ctx.String(ServiceConfigPathFlagName),
		TLSConfig:         optls.ReadCLIConfig(ctx),
		RPCConfig:         oprpc.ReadCLIConfig(ctx),
		LogConfig:         oplog.ReadCLIConfig(ctx),
		MetricsConfig:     opmetrics.ReadCLIConfig(ctx),
		PprofConfig:       oppprof.ReadCLIConfig(ctx),
		HealthzConfig:     ReadHealthzCLIConfig(ctx),
	}
}
