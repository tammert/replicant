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

	rootCmd.PersistentFlags().StringP("config", "c", "/config/replicant.yaml", "File containing the configuration for Replicant")
	rootCmd.PersistentFlags().BoolP("replace-tag", "r", false, "Replace images with the same tag, if the image SHA is different")
	rootCmd.PersistentFlags().BoolP("allow-prerelease", "p", false, "Include prerelease versions (e.g. 1.2.3-alpha1) when mirroring")
	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("replace-tag", rootCmd.PersistentFlags().Lookup("replace-tag"))
	viper.BindPFlag("allow-prerelease", rootCmd.PersistentFlags().Lookup("allow-prerelease"))

	viper.AutomaticEnv()

	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}
