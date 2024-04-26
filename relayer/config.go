package relayer

import (
	"fmt"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

type Config struct {
	Chains map[string]ChainConfig `yaml:"chains"` // map of chain-name to a chain configuration
}

const configFileName = "config.yaml"

func GetConfigPath(dirPath string) string {
	return filepath.Join(dirPath, configFileName)
}

type ChainConfig struct {
	ChainID        string  `yaml:"chain-id"`
	RPCAddr        string  `yaml:"rpc-addr"`
	AccountPrefix  string  `yaml:"account-prefix"`
	GasAdjustment  float64 `yaml:"gas-adjustment"`
	GasPrices      string  `yaml:"gas-prices"`
	Gas            string  `yaml:"gas"`
	KeyringBackend string  `yaml:"keyring-backend"`
	KeyName        string  `yaml:"key-name"`
}

func (config Config) Validate() error {
	for chainName, chainConfig := range config.Chains {
		if chainConfig.ChainID == "" {
			return fmt.Errorf("chain ID is required for chain %s", chainName)
		}

		if chainConfig.RPCAddr == "" {
			return fmt.Errorf("rpc-address is required for chain %s", chainName)
		}

		if chainConfig.AccountPrefix == "" {
			return fmt.Errorf("account-prefix is required for chain %s", chainName)
		}

		if chainConfig.GasAdjustment == 0 {
			return fmt.Errorf("gas-adjustment is required for chain %s", chainName)
		}

		if chainConfig.GasPrices == "" {
			return fmt.Errorf("gas-prices is required for chain %s", chainName)
		}

		if chainConfig.Gas == "" {
			return fmt.Errorf("gas is required for chain %s", chainName)
		}

		if chainConfig.KeyringBackend == "" {
			return fmt.Errorf("keyring-backend is required for chain %s", chainName)
		}

		if chainConfig.KeyName == "" {
			return fmt.Errorf("key-name is required for chain %s", chainName)
		}
	}

	return nil
}

func ConfigExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func ReadConfigFromFile(logger *zap.Logger, path string) (Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var config Config
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return Config{}, err
	}

	logger.Debug("Read config from file", zap.String("path", path), zap.Any("config", config))

	return config, nil
}

func WriteConfigToFile(config Config, path string) error {
	// Make sure there is no config file at the path
	if ConfigExists(path) {
		return fmt.Errorf("config file already exists at %s", path)
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
