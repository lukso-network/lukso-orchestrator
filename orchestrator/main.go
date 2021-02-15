package main

import (
	"os"
	"fmt"

	"github.com/urfave/cli/v2"

)

func main() {
	app := &cli.App{
		Name: "orchestrator",
		Usage: "Lukso orchestrator implementation for lukso-mainnet 1.0",
		Action: func(c *cli.Context) error {
			//log.Info("Starting Lusko Orchestrator Client!!!")
			fmt.Println("This is test")
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		//log.Fatal(err)
	}
}