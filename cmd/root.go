package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	replicant "replicant/internal"
)

var rootCmd = &cobra.Command{
	Use:   "replicant",
	Short: "Replicant mirrors public container images to private registries",
	Run: func(cmd *cobra.Command, args []string) {
		replicant.CloneToRepo()
		replicant.ListTags()
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

	rootCmd.PersistentFlags().StringP("image", "i", "ratelimitalways/test:latest", "Container image to mirror")
	rootCmd.PersistentFlags().StringP("repository", "r", "ratelimitalways/test", "Container repository to list tags")
	viper.BindPFlag("image", rootCmd.PersistentFlags().Lookup("image"))
	viper.BindPFlag("repository", rootCmd.PersistentFlags().Lookup("repository"))

	viper.AutomaticEnv()
}
