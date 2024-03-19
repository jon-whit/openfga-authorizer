package config

import "fmt"

type Config struct {
	HostPortAddr     string `json:"hostPortAddr"`
	HTTPSCertDir     string `json:"httpsCertDir"`
	HTTPSCertName    string `json:"httpsCertName"`
	HTTPSCertKeyName string `json:"httpsCertKeyName"`
	HTTPSCertCAName  string `json:"httpsCertCAName"`

	HealthProbeAddr      string `json:"healthProbeAddr"`
	EnableLeaderElection bool   `json:"enableLeaderElection"`

	FGAHostPortAddr string `json:"fgaAddr"`
	FGAStoreID      string `json:"fgaStoreID"`
	FGAClientID     string `json:"fgaClientID"`
	FGAClientSecret string `json:"fgaClientSecret"`
	FGAAudience     string `json:"fgaAudience"`
	FGATokenIssuer  string `json:"fgaTokenIssuer"`
}

func DefaultConfig() *Config {
	cfg := &Config{
		HostPortAddr:         ":9443",
		HealthProbeAddr:      ":9002",
		HTTPSCertDir:         "certs",
		HTTPSCertName:        "server.crt",
		HTTPSCertKeyName:     "server.key",
		HTTPSCertCAName:      "ca.crt",
		EnableLeaderElection: false,
		FGAHostPortAddr:      "openfga:8081", // defaults to default gRPC OpenFGA port
	}

	return cfg
}

func (c *Config) Validate() error {
	if c.FGAStoreID == "" {
		return fmt.Errorf("'fgaStoreID' config must be set")
	}

	return nil
}
