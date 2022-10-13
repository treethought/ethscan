package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/treethought/ethscan/ui"

	"github.com/spf13/viper"
)

var (
	cfgFile string
	config  *ui.Config
)

var rootCmd = &cobra.Command{
	Use:   "ethscan",
	Short: "Ethereum block explorer for the terminal",
	Run: func(cmd *cobra.Command, args []string) {
		app := ui.NewApp(config)
		app.Init()

		if len(args) == 1 {
			app.StartWith(args[0])
			return
		}

		app.Start()

	},
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/ethscan.yaml)")
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

		// Search config in home directory with name ".cli" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("ethscan")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

	conf := &ui.Config{
		RpcUrl:       viper.GetString("rpc_url"),
		EtherscanKey: viper.GetString("etherscan_key"),
		DisableENS:   viper.GetBool("disable_ens"),
	}

	config = conf

}
