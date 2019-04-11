package main

import (
	"os"

	"github.com/kckecheng/powermaxbeat/cmd"

	_ "github.com/kckecheng/powermaxbeat/include"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
