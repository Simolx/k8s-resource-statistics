/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/util/homedir"
)

const (
	KUBECONFIGKEY = "kubeconfig"
)

var cfgFile string

var clientset *kubernetes.Clientset

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "k8s",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig, initClient, initLoggingFlags)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	if home := homedir.HomeDir(); home != "" {
		rootCmd.PersistentFlags().String(KUBECONFIGKEY, filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		rootCmd.PersistentFlags().String(KUBECONFIGKEY, "", "absolute path to the kubeconfig file")
	}
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.k8s.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
}

func initClient() {
	if config, err := clientcmd.BuildConfigFromFlags("", viper.GetString(KUBECONFIGKEY)); err != nil {
		cobra.CheckErr(err)
	} else {
		clientset, err = kubernetes.NewForConfig(config)
		cobra.CheckErr(err)
	}
}

func initLoggingFlags() {
	klog.InitFlags(nil)
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

		// Search config in home directory with name ".k8s" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".k8s")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil && !errors.As(err, &viper.ConfigFileNotFoundError{}) {
		cobra.CheckErr(err)
	}
	kubeConfigFlag := rootCmd.PersistentFlags().Lookup(KUBECONFIGKEY)
	if fileInfo, err := os.Stat(kubeConfigFlag.Value.String()); err == nil && !fileInfo.IsDir() {
		viper.BindPFlag(KUBECONFIGKEY, kubeConfigFlag)
	}
}
