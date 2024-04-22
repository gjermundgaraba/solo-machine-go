package relayer

import (
	"fmt"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

type Config struct {
	configPath  string            `yaml:"-"`
	SoloMachine SoloMachineConfig `yaml:"solo-machine"`
	CosmosChain CosmosChainConfig `yaml:"cosmos-chain"`
}

type SoloMachineConfig struct {
}

type CosmosChainConfig struct {
	RPCAddr        string  `yaml:"rpc-addr"`
	AccountPrefix  string  `yaml:"account-prefix"`
	ChainID        string  `yaml:"chain-id"`
	GasAdjustment  float64 `yaml:"gas-adjustment"`
	GasPrices      string  `yaml:"gas-prices"`
	Gas            string  `yaml:"gas"`
	KeyringBackend string  `yaml:"keyring-backend"`
	Key            string  `yaml:"key"`

	SoloMachineLightClient SoloMachineLightClientConfig `yaml:"solo-machine-light-client"`
}

type SoloMachineLightClientConfig struct {
	IBCClientID  string `yaml:"ibc-client-id"`
	ConnectionID string `yaml:"connection-id"`
	ChannelID    string `yaml:"channel-id"`
}

// TODO: Add validation method to check that all light client stuff is present

func (config *Config) Validate() error {
	if config.CosmosChain.RPCAddr == "" {
		return fmt.Errorf("blockchain rpc address is required")
	}

	if config.CosmosChain.AccountPrefix == "" {
		return fmt.Errorf("blockchain account prefix is required")
	}

	if config.CosmosChain.ChainID == "" {
		return fmt.Errorf("blockchain chain ID is required")
	}

	if config.CosmosChain.GasAdjustment == 0 {
		return fmt.Errorf("blockchain gas adjustment is required")
	}

	if config.CosmosChain.GasPrices == "" {
		return fmt.Errorf("blockchain gas prices is required")
	}

	if config.CosmosChain.KeyringBackend == "" {
		return fmt.Errorf("blockchain keyring backend is required")
	}

	if config.CosmosChain.Key == "" {
		return fmt.Errorf("blockchain key is required")
	}

	return nil
}

func ReadConfigFromFile(logger *zap.Logger, path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}
	config.configPath = path

	logger.Debug("Read config from file", zap.String("path", path), zap.Any("config", config))

	return &config, nil
}

// If path is set to "", the config will be written to the path it was read from
func WriteConfigToFile(config *Config, path string, force bool) error {
	if path == "" {
		path = config.configPath
		if path == "" {
			return fmt.Errorf("no path specified and config has no path")
		}
	}

	if !force {
		// Make sure there is no config file at the path
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("config file already exists at %s", path)
		}
	}

	// Make sure the full directory path exists
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	file, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(path, file, 0644)
}
