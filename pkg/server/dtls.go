package server

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/aura-speak/networking/internal/config"
	"github.com/pion/dtls/v3"
)

func NewDTLSServerMTLConfig(cfg *config.ServerConfig) (*dtls.Config, error) {
	cert, err := tls.LoadX509KeyPair(
		fmt.Sprintf("%s/%s", cfg.Server.DTLS.Path, cfg.Server.DTLS.Cert),
		fmt.Sprintf("%s/%s", cfg.Server.DTLS.Path, cfg.Server.DTLS.Key),
	)
	if err != nil {
		return nil, err
	}

	caPam, err := os.ReadFile(fmt.Sprintf("%s/%s", cfg.Server.DTLS.Path, cfg.Server.DTLS.CA))
	if err != nil {
		return nil, err
	}
	clientCAs := x509.NewCertPool()
	if !clientCAs.AppendCertsFromPEM(caPam) {
		return nil, err
	}

	return &dtls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   dtls.RequireAndVerifyClientCert,
		ClientCAs:    clientCAs,

		CipherSuites: []dtls.CipherSuiteID{
			dtls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			dtls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			dtls.TLS_ECDHE_ECDSA_WITH_AES_128_CCM,
		},

		MTU: 1200,
	}, nil
}
