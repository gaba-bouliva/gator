package main

import (
	"context"
	"fmt"

	"github.com/gaba-bouliva/gator/internal/application"
	"github.com/gaba-bouliva/gator/internal/database"
)

func middlewareLoggedIn(
	handler func(
		app *application.App,
		cmd application.Command,
		user database.User) error) func(*application.App, application.Command) error {

	return func(a *application.App, c application.Command) error {
		username, err := a.Config.GetCurrentUser()
		if err != nil {
			fmt.Println("login failed error encountered")
			return err
		}

		usr, err := a.DB.GetUserByName(context.Background(), username)
		if err != nil {
			fmt.Println("error user not found")
			return err
		}

		return handler(a, c, usr)
	}

}
