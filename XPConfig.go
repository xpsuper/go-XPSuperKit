package XPSuperKit

import (
	"os"
	"regexp"
)

type XPConfigImpl struct {
	*XPConfigEnvironment
}

type XPConfigEnvironment struct {
	Environment       string
	EnvironmentPrefix string
}

func NewXPConfig(configEnv *XPConfigEnvironment) *XPConfigImpl {
	if configEnv == nil {
		configEnv = &XPConfigEnvironment{}
	}
	return &XPConfigImpl{XPConfigEnvironment: configEnv}
}

func (cfg *XPConfigImpl) GetEnvironment() string {
	if cfg.Environment == "" {
		if env := os.Getenv("XPCONFIG_ENV"); env != "" {
			return env
		}

		if isTest, _ := regexp.MatchString("/_test/", os.Args[0]); isTest {
			return "test"
		}

		return "development"
	}

	return cfg.Environment
}

func (cfg *XPConfigImpl) Load(config interface{}, files ...string) error {
	for _, file := range cfg.getConfigurationFiles(files...) {
		if err := processFile(config, file); err != nil {
			return err
		}
	}

	if prefix := cfg.getEnvironmentPrefix(config); prefix == "-" {
		return processTags(config)
	} else {
		return processTags(config, prefix)
	}
}
