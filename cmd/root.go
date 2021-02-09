package cmd

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	replicant "replicant/internal"
)

var rootCmd = &cobra.Command{
	Use:   "replicant",
	Short: "Replicant mirrors public container images to private registries",
	Run: func(cmd *cobra.Command, args []string) {
		replicant.Run(viper.GetString("config"))
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	viper.SetEnvPrefix("replicant")

	rootCmd.PersistentFlags().StringP("config", "c", "/home/tammert/github/tammert/replicant/config.yaml", "File containing the configuration for Replicant")
	rootCmd.PersistentFlags().BoolP("replace-tag", "r", false, "If you want to override downstream with same tag but different SHA")
	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("replace-tag", rootCmd.PersistentFlags().Lookup("replace-tag"))

	viper.AutomaticEnv()

	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}
