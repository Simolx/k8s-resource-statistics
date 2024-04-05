/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"example.com/dev/k8s/cmd"
)

type mm struct {
	Name string `json:"name,omitempty"`
}

func main() {
	cmd.Execute()
}
