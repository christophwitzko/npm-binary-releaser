package main

import (
	"fmt"
	"log"
	"os"

	"github.com/christophwitzko/npm-binary-releaser/pkg/releaser"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var VERSION string

func main() {
	cmd := &cobra.Command{
		Use:     "npm-binary-releaser",
		Short:   "npm-binary-releaser - release binaries to npm",
		Run:     cliHandler,
		Version: VERSION,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}
	SetFlags(cmd)

	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Print the current npm-binary-releaser config",
		Run: func(cmd *cobra.Command, args []string) {
			shouldValidate, _ := cmd.Flags().GetBool("validate")
			c := NewConfig(cmd)
			if shouldValidate {
				if err := c.Validate(); err != nil {
					fmt.Printf("config validation error: %s\n", err)
					os.Exit(1)
				}
			}
			cfgStr, _ := yaml.Marshal(c)
			fmt.Printf("# .npm-binary-releaser.yaml\n%s", string(cfgStr))
		},
	}
	configCmd.Flags().Bool("validate", false, "validate the config")
	cmd.AddCommand(configCmd)

	cobra.OnInitialize(func() {
		if err := InitConfig(); err != nil {
			fmt.Printf("\nConfig error: %s\n", err.Error())
			os.Exit(1)
		}
	})

	if err := cmd.Execute(); err != nil {
		fmt.Printf("\n%s\n", err.Error())
		os.Exit(1)
	}
}

func cliHandler(cmd *cobra.Command, args []string) {
	var logger = log.New(os.Stderr, "[npm-binary-releaser]: ", 0)
	if err := releaser.Run(NewConfig(cmd), logger); err != nil {
		logger.Println(err)
		os.Exit(1)
		return
	}
}
