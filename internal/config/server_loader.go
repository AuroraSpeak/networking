package config

import (
	"os"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// ServerConfigLoader loads the server configuration from "server_config.yml".
// If the file does not exist, it writes a sensible default using
// `WriteDefaultServerConfig` and then reloads it. The function returns a
// pointer to the decoded `ServerConfig`. For unrecoverable I/O errors it
// logs the error and exits the process.
func ServerConfigLoader() *ServerConfig {
	const cfgPath = "server_config.yml"

	f, err := os.Open(cfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Infof("%s not found: writing default config", cfgPath)
			if werr := WriteDefaultServerConfig(cfgPath); werr != nil {
				log.WithError(werr).Error("Cant write default Server Config")
				os.Exit(1)
			}
			f, err = os.Open(cfgPath)
			if err != nil {
				log.WithError(err).Error("Cant read Server Config after creating default")
				os.Exit(1)
			}
		} else {
			log.WithError(err).Error("Cant read Server Config")
			os.Exit(1)
		}
	}
	defer f.Close()

	var config ServerConfig
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&config)
	if err != nil {
		log.WithError(err).Error("Cant decode yaml")
	}
	return &config
}

// WriteDefaultServerConfig writes a default server_config.yml to the given path.
func WriteDefaultServerConfig(path string) error {
	cfg := ServerConfig{}
	cfg.Server.Port = "8080"
	cfg.Server.Host = "0.0.0.0"
	cfg.Server.DTLS.Path = "certs/"
	cfg.Server.DTLS.Cert = ""
	cfg.Server.DTLS.Key = ""
	cfg.Server.DTLS.CA = ""

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := yaml.NewEncoder(f)
	defer encoder.Close()
	if err := encoder.Encode(&cfg); err != nil {
		return err
	}
	return nil
}
