package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gaba-bouliva/gator/internal/application"
	"github.com/gaba-bouliva/gator/internal/config"
	"github.com/gaba-bouliva/gator/internal/database"
	"github.com/google/uuid"

	_ "github.com/lib/pq"
)

var app = application.NewApp(nil)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println("error reading config")
		log.Fatalln(err)
	}
	app.Config = cfg
	db, err := sql.Open("postgres", app.Config.DBUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	app = application.NewApp(db)

	app.RegisterCMD("login", handleLogin)
	app.RegisterCMD("register", handleRegister)

	args := os.Args

	if len(args) < 2 {
		fmt.Println("invalid syntax")
		log.Fatal("[usage] command-name <arguments>")
	}

	cmdName := args[1]
	cmdArgs := args[2:]

	newCmd := application.Command{
		Name:      cmdName,
		Arguments: cmdArgs,
	}

	err = app.RunCMD(newCmd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func handleLogin(a *application.App, cmd application.Command) error {
	if len(cmd.Arguments) < 1 {
		log.Fatal("login command expects a single argument, username")
	}

	if len(strings.Trim(cmd.Arguments[0], " ")) < 1 {
		log.Fatal("valid username required")
	}

	userName := cmd.Arguments[0]

	user, err := a.DB.GetUser(context.Background(), userName)
	if err != nil {
		return err
	}

	err = a.Config.SetUser(user.Name)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("username login successful")
	return nil
}

func handleRegister(a *application.App, cmd application.Command) error {
	if len(cmd.Arguments) < 1 {
		log.Fatal("login command expects a single argument, username")
	}

	if len(strings.Trim(cmd.Arguments[0], " ")) < 1 {
		log.Fatal("valid username required")
	}
	createUserParams := database.CreateUserParams{
		ID:        int32(uuid.New().ID()),
		Name:      cmd.Arguments[0],
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	newUser, err := a.DB.CreateUser(context.Background(), createUserParams)
	if err != nil {
		return err
	}

	err = a.Config.SetUser(newUser.Name)
	if err != nil {
		return err
	}
	fmt.Printf("New user registered: %+v", newUser)

	return nil
}
