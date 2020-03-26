// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import "time"

type Config struct {
	Period           time.Duration `config:"period"`
	Secret           string        `config:"client_secret"`
	ClientID         string        `config:"client_id"`
	TenantID         string        `config:"tenant_id"`
	RegistryFilePath string        `config:"registry_file_path"`
	StartDate        string        `config:"start_date"`
}

var DefaultConfig = Config{
	Period:           10 * time.Minute,
	RegistryFilePath: "./msgraph.state",
	//get Last 90 days
	StartDate: time.Now().Add(-(7 * 24) * time.Hour).UTC().Format(time.RFC3339),
}
