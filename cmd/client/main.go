package main

import (
	"fmt"
	"os"

	"github.com/biandc/nhole/client"
	"github.com/biandc/nhole/pkg/config"
	ver "github.com/biandc/nhole/pkg/version"
	"github.com/spf13/cobra"
)

const (
	NHOLETYPE = "nhole-client"
)

var (
	version bool
	cfgFile string
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&version, "version", "v", false, fmt.Sprintf("%s version.", NHOLETYPE))
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "cfg_file", "c", fmt.Sprintf("./%s.yaml", NHOLETYPE), "config file path.")
	rootCmd.PersistentFlags().StringVarP(&client.LogWay, "log_way", "", "console", "log way.(console|file)")
	rootCmd.PersistentFlags().StringVarP(&client.LogFile, "log_file", "", "", "log save file.")
	rootCmd.PersistentFlags().StringVarP(&client.LogLevel, "log_level", "", "info", "log level.(error|warn|info|debug|trace)")
	rootCmd.PersistentFlags().BoolVarP(&client.LogDisableColor, "log_disable_color", "", false, "disable log color.")
}

var rootCmd = &cobra.Command{
	Use: NHOLETYPE,
	RunE: func(cmd *cobra.Command, args []string) error {
		if version {
			ver.ShowVersion()
			return nil
		}
		cfg, err := config.UnmarshalClientCfgByFile(cfgFile)
		if err != nil {
			return err
		}
		err = client.Run(cfg)
		return err
	},
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
