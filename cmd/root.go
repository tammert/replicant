package cmd

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	replicant "replicant/internal"
	"strings"
)

var rootCmd = &cobra.Command{
	Use:   "replicant",
	Short: "Replicant mirrors container images between repositories",
	Run: func(cmd *cobra.Command, args []string) {
		// From here on out flags/environment are parsed.
		setupLogging()
		replicant.Run(viper.GetString("config-file"))
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// cobra
	rootCmd.PersistentFlags().StringP("config-file", "c", "/config/replicant.yaml", "File containing the configuration for Replicant")
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enables debug logging")

	// viper
	viper.SetEnvPrefix("replicant")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	err := viper.BindPFlags(rootCmd.PersistentFlags())
	if err != nil {
		fmt.Println("error binding flags, exiting")
		os.Exit(1)
	}
	viper.AutomaticEnv()
}

func setupLogging() {
	if viper.GetBool("debug") {
		log.SetLevel(log.DebugLevel)
	}
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
}
