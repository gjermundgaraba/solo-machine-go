package main

import (
	"github.com/gjermundgaraba/solo-machine-go/cmd"
	"github.com/gjermundgaraba/solo-machine-go/utils"
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
