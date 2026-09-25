package stytch

import platform "github.com/moasq/go-b2b-starter/internal/platform/stytch"

const (
	EnvTest = platform.EnvTest
	EnvLive = platform.EnvLive
)

// Config shares one environment contract with member and organization operations.
type Config platform.Config

func LoadConfig() (*Config, error) { cfg, err := platform.LoadConfig(); return (*Config)(cfg), err }
func (c *Config) Validate() error  { return (*platform.Config)(c).Validate() }
func (c *Config) Configured() bool { return (*platform.Config)(c).Configured() }
