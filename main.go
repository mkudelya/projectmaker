package main

import (
	"fmt"
	"github.com/mkudelya/projectmaker/internal/app"
	"os"
)

func main() {
	app := app.NewApp()
	err := app.InitApp(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = app.ExecuteCommand()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}
