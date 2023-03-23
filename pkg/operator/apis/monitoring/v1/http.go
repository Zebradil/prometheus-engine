package v1

import (
	"github.com/prometheus/common/config"
)

func (c *TLSConfig) ToPrometheusConfig() *config.TLSConfig {
	return &config.TLSConfig{
		InsecureSkipVerify: c.InsecureSkipVerify,
		ServerName:         c.ServerName,
	}
}
