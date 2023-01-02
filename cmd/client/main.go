package main

import (
	"fmt"
	ver "github.com/biandc/nhole/pkg/version"
	"os"

	"github.com/biandc/nhole/client"
	"github.com/biandc/nhole/pkg/config"
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
