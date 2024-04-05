/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"example.com/dev/k8s/controllers"
	"example.com/dev/k8s/utils"
	"k8s.io/klog/v2"

	"github.com/spf13/cobra"
)

// resourceCmd represents the resource command

var requestNamespaces []string
var jsonFile, csvFile, excelFile string

var resourceCmd = &cobra.Command{
	Use:   "resource",
	Short: "Get k8s resources",
	Long:  `Get k8s resources: namespace, deployment, statefulset`,
	Run: func(cmd *cobra.Command, args []string) {
		var namespaces []string
		var err error
		if len(requestNamespaces) > 0 {
			namespaces = requestNamespaces
		} else {
			namespaces, err = controllers.GetNamespaces(clientset)
			cobra.CheckErr(err)
		}
		klog.Infof("requests namespace %#v", namespaces)
		result, err := controllers.GetControllerItems(clientset, namespaces)
		cobra.CheckErr(err)
		if len(jsonFile) > 0 {
			cobra.CheckErr(
				utils.WriteJsonFile(
					struct {
						Responses []controllers.ControllerItem `json:"responses,omitempty"`
					}{
						result,
					},
					jsonFile))
		}
		if len(csvFile) > 0 {
			cobra.CheckErr(utils.WriteCsvFile(controllers.ConvertResultToCsv(result), nil, csvFile))
		}
		if len(excelFile) > 0 {
			cobra.CheckErr(utils.WriteExcelFile(result, excelFile, "resources"))
		}
	},
}

func init() {
	rootCmd.AddCommand(resourceCmd)

	resourceCmd.Flags().StringArrayVarP(&requestNamespaces, "namespace", "n", []string{}, "specified namespace")

	resourceCmd.Flags().StringVar(&jsonFile, "json", "", "json file path for result")

	resourceCmd.Flags().StringVar(&csvFile, "csv", "", "csv file path for result")

	resourceCmd.Flags().StringVar(&excelFile, "excel", "", "excel file path for result")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// resourceCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
}
