package config

// ClientExt returns optional client config values by namespace.
func ClientExt(c *Config, t ClientType) Values {
	configs := Ext()
	result := make(Values, len(configs))

	for _, conf := range configs {
		result[conf.name] = conf.clientValues(c, t)
	}

	return result
}
