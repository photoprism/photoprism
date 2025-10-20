package config

// ClientExt collects optional client config values provided by registered extensions.
func ClientExt(c *Config, t ClientType) Values {
	bootConfigs := Ext(StageBoot)
	initConfigs := Ext(StageInit)

	result := make(Values, len(bootConfigs)+len(initConfigs))

	for _, conf := range bootConfigs {
		if conf.clientValues == nil {
			continue
		}
		result[conf.name] = conf.clientValues(c, t)
	}

	for _, conf := range initConfigs {
		if conf.clientValues == nil {
			continue
		}
		result[conf.name] = conf.clientValues(c, t)
	}

	return result
}
