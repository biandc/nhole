package main

import (
	"fmt"
	"os"

	"github.com/biandc/nhole/pkg/config"
	ver "github.com/biandc/nhole/pkg/version"
	"github.com/biandc/nhole/server"
	"github.com/spf13/cobra"
)

const (
	NHOLETYPE = "nhole-server"
)

var (
	version bool
	cfgFile string
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&version, "version", "v", false, fmt.Sprintf("%s version.", NHOLETYPE))
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "cfg_file", "c", fmt.Sprintf("./%s.yaml", NHOLETYPE), "config file path.")
}

var rootCmd = &cobra.Command{
	Use: NHOLETYPE,
	RunE: func(cmd *cobra.Command, args []string) error {
		if version {
			ver.ShowVersion()
			return nil
		}
		cfg, err := config.UnmarshalServerCfgByFile(cfgFile)
		if err != nil {
			return err
		}
		err = server.Run(cfg)
		return err
	},
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
