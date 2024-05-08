package types

// Config structure for application configuration
type Config struct {
	Port    int    `yaml:"port"`
	DataDir string `yaml:"data_dir"`
	AESKey  string `yaml:"aes_key"`
}
