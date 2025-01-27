package application

import (
	"database/sql"
	"fmt"

	"github.com/gaba-bouliva/gator/internal/config"
	"github.com/gaba-bouliva/gator/internal/database"
)

type App struct {
	Config   *config.Config
	Commands map[string]func(*App, Command) error
	DB       *database.Queries
}

func NewApp(db *sql.DB) *App {
	return &App{
		Config:   &config.Config{},
		Commands: make(map[string]func(*App, Command) error),
		DB:       database.New(db),
	}
}

func (a *App) RegisterCMD(name string, f func(*App, Command) error) {
	a.Commands[name] = f
}

func (a *App) RunCMD(cmd Command) error {
	handler, exists := a.Commands[cmd.Name]
	if !exists {
		return fmt.Errorf("invalid command provided")
	}
	return handler(a, cmd)
}
