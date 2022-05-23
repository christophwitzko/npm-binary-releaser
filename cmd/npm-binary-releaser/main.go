package main

import (
	"fmt"
	"log"
	"os"

	"github.com/christophwitzko/npm-binary-releaser/pkg/config"
	"github.com/christophwitzko/npm-binary-releaser/pkg/releaser"
	"github.com/spf13/cobra"
)

var VERSION string

func main() {
	cmd := &cobra.Command{
		Use:     "npm-binary-releaser",
		Short:   "npm-binary-releaser - release binaries to npm",
		Run:     cliHandler,
		Version: VERSION,
	}

	config.InitConfig(cmd)

	if err := cmd.Execute(); err != nil {
		fmt.Printf("\n%s\n", err.Error())
		os.Exit(1)
	}
}

func cliHandler(cmd *cobra.Command, args []string) {
	var logger = log.New(os.Stderr, "[npm-binary-releaser]: ", 0)
	if err := releaser.Run(config.NewConfig(cmd), logger); err != nil {
		logger.Println(err)
		os.Exit(1)
		return
	}
}
