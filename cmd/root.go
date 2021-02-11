// Package cmd holds the cobra setup for CLI commands/subcommands
package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/apeunit/LaunchControlD/pkg/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "lctrld",
	Short: "The LaunchControl command & control service",
	Long:  ``,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(version string) {
	rootCmd.Version = version
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var (
	debug    = false
	settings = new(config.Schema)
)

func init() {
	cobra.OnInitialize(initConfig)
	// config file
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/lctrld/config.yaml)")
	// verbose logging
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Enable debug logging")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if debug {
		// log.SetReportCaller(true)
		// Output to stdout instead of the default stderr
		// Can be any io.Writer, see below for File example
		// log.SetOutput(os.Stdout)

		// Only log the warning severity or above.
		log.SetLevel(log.DebugLevel)
	}
	fmt.Printf(`
┌─┐┬  ┬┌┬┐┬  ┬┌─┐╔╦╗
├┤ └┐┌┘ │ └┐┌┘┌─┘ ║║
└─┘ └┘  ┴  └┘ └─┘═╩╝ %s`, rootCmd.Version)
	fmt.Println()

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in home directory with name ".LaunchControlD" (without extension).
		viper.AddConfigPath(".")
		viper.AddConfigPath("/etc/lctrld")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match
	config.Defaults()
	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Error loading config file:", viper.ConfigFileUsed(), ":", err)
	} else {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
		settingsParseError := viper.UnmarshalExact(&settings)
		if settingsParseError != nil {
			log.Debugf("Errors encountered while parsing %s: %s", viper.ConfigFileUsed(), settingsParseError)
		}
	}
	// set runtime data
	settings.RuntimeVersion = rootCmd.Version
	settings.RuntimeStartedAt = time.Now()
}
