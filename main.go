package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gaba-bouliva/gator/internal/application"
	"github.com/gaba-bouliva/gator/internal/config"
)

var app = application.NewApp()

func main() { 
	cfg, err := config.Read()
	if err != nil {
		log.Fatalln(err)
	}

	app.Config = cfg

	app.RegisterCMD("login", handleLogin)

	args := os.Args

	if len(args) < 2 { 
		fmt.Println("invalid syntax")
		log.Fatal("[usage] command-name <arguments>")
	}

	cmdName := args[1]
	cmdArgs := args[2:]
	
	newCmd := application.Command{
		Name: cmdName,
		Arguments: cmdArgs,
	}

	fmt.Printf("new command: %+v\n", newCmd)
	err = app.RunCMD(newCmd)
	if err != nil {
		log.Fatal(err)
	}
}

func handleLogin(a *application.App, cmd application.Command) error {
	if len(cmd.Arguments) < 1 { 
		log.Fatal("login command expects a single argument, username")
	}	
	
	if len(strings.Trim(cmd.Arguments[0], " ")) < 1 {
		log.Fatal("valid username required")
	}

	err := a.Config.SetUser(cmd.Arguments[0])
	if err != nil { 
		log.Fatal(err)
	}
	
	fmt.Println("username has be set")
	return nil
}
