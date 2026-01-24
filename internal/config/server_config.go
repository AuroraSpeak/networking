package config

type ServerConfig struct {
	Server struct {
		Port string `yaml:"port"`
		// Implement Later
		Host string `yaml:"host"`
		DTLS struct {
			Path string `yaml:"path"`
			Cert string `yaml:"cert"`
			Key  string `yaml:"key"`
			CA   string `yaml:"ca"`
		} `yaml:"dtls"`
		Env string `yaml:"env"` // prod or env if nothing is set then prod
	} `yaml:"server"`
}
