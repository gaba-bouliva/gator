package application

import (
	"fmt"

	"github.com/gaba-bouliva/gator/internal/config"
)


type App struct { 
	Config *config.Config
	Commands map[string]func(*App, Command) error
}

func NewApp() *App { 
	return &App{
		Config: &config.Config{},
		Commands: make(map[string]func(*App, Command) error),
	}
}


func (a *App) RegisterCMD(name string, f func (*App, Command)error) { 
		a.Commands[name] = f	
}

func (a *App) RunCMD(cmd Command) error { 
	handler, exists := a.Commands[cmd.Name]
	if !exists { 
		return fmt.Errorf("invalid command provided")
	}
	return handler(a,cmd)
}
