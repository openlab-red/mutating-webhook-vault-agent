package cmd

import (
	"github.com/openlab-red/mutating-webhook-vault-agent/internal/engine"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var handlerCmd = &cobra.Command{
	Use:   "start",
	Short: "Start webhook server",
	Long:  `Start webhook server`,
	Run: func(cmd *cobra.Command, args []string) {
		engine.Start()
	},
}

func init() {
	RootCmd.AddCommand(handlerCmd)
	viper.SetDefault("log-level", "INFO")
	viper.SetDefault("port", "8080")
}
