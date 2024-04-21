package main

import "github.com/gjermundgaraba/solo-machine-go/server/cmd"

func main() {
	rootCmd := cmd.RootCmd()

	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
}
