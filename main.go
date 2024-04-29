package main

import (
	"github.com/gjermundgaraba/solo-machine/cmd"
	"github.com/gjermundgaraba/solo-machine/utils"
	"go.uber.org/zap"
)

func main() {
	rootCmd := cmd.RootCmd()

	err := rootCmd.Execute()
	if err != nil {
		utils.CreateLogger(false).Error("Error caught", zap.Error(err))
		panic(err)
	}
}
