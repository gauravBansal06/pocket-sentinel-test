package config

import "github.com/spf13/viper"

func setDefaultConfig() {
	viper.SetDefault("LogConfig.EnableConsole", true)
	viper.SetDefault("LogConfig.ConsoleJSONFormat", false)
	viper.SetDefault("LogConfig.ConsoleLevel", "info")
	viper.SetDefault("LogConfig.EnableFile", true)
	viper.SetDefault("LogConfig.FileJSONFormat", true)
	viper.SetDefault("LogConfig.FileLevel", "debug")
	viper.SetDefault("LogConfig.FileLocation", "./exemplar.log")
	viper.SetDefault("Env", "prod")
	viper.SetDefault("Port", "9876")
	viper.SetDefault("GRPCPort", "12000")
	viper.SetDefault("Verbose", false)
}
