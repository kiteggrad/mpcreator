/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"go.uber.org/zap"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mpcreator",
	Short: "инструмент для быстрого клонирования всех репозиториев из gitlab",
	Long: `mpcreator - это инструмент для быстрого клонирования всех репозиториев из gitlab
в один главный репозиторий (главный проект - mp).`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(func() {
		initConfig()

		logLevel, err := rootCmd.PersistentFlags().GetString("loglevel")
		if err != nil {
			panic(err)
		}

		initLogger(logLevel)
	})

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mpcreator.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.PersistentFlags().String("loglevel", "info", "log level")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".mpcreator" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".mpcreator")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func initLogger(level string) *zap.Logger {
	var err error

	zapCfg := zap.NewDevelopmentConfig()
	zapCfg.Development = false // do not show trace for warn level
	zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	if level != "" {
		zapCfg.Level, err = zap.ParseAtomicLevel(level)
		if err != nil {
			err = fmt.Errorf("failed to zap.ParseAtomicLevel: %w", err)
			err = fmt.Errorf("failed to initLogger: %w", err)
			panic(err)
		}
	}

	logger, err := zapCfg.Build()
	if err != nil {
		err = fmt.Errorf("failed to zapCfg.Build: %w", err)
		err = fmt.Errorf("failed to initLogger: %w", err)
		panic(err)
	}

	zap.ReplaceGlobals(logger)

	return logger
}
