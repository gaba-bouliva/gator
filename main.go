package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
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
	app.RegisterCMD("reset", handleReset)
	app.RegisterCMD("users", handleUsers)
	app.RegisterCMD("agg", handleAgg)
	app.RegisterCMD("addfeed", handleAddFeed)
	app.RegisterCMD("feeds", handleFeeds)
	app.RegisterCMD("follow", handleFollow)
	app.RegisterCMD("following", handleFollowing)

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

func handleFollowing(a *application.App, cmd application.Command) error {
	username, err := a.Config.GetCurrentUser()
	if err != nil {
		return err
	}

	user, err := a.DB.GetUserByName(context.Background(), username)
	if err != nil {
		return err
	}

	usrFeedFollowings, err := a.DB.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return err
	}

	for _, feedsFollow := range usrFeedFollowings {
		fmt.Println("feed name: ", feedsFollow.FeedName)
	}

	return nil
}

func handleFollow(a *application.App, cmd application.Command) error {
	nbrArgs := 1
	err := checkCMDArgs(cmd, nbrArgs)
	if err != nil {
		return err
	}

	username, err := a.Config.GetCurrentUser()
	if err != nil {
		return err
	}

	user, err := a.DB.GetUserByName(context.Background(), username)
	if err != nil {
		return err
	}

	feed, err := a.DB.GetFeedByURL(context.Background(), cmd.Arguments[0])
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return fmt.Errorf("feed not found with url %s", cmd.Arguments[0])
		} else {
			return err
		}
	}

	createdFeedFollowParam := database.CreateFeedFollowParams{
		ID:        int32(uuid.New().ID()),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedsID:   feed.ID,
	}
	feed_follow, err := a.DB.CreateFeedFollow(context.Background(), createdFeedFollowParam)
	if err != nil {
		fmt.Println("error creating feed_follow")
		return err
	}

	fmt.Println(feed_follow.FeedName)
	fmt.Println(feed_follow.UserName)

	return nil
}

func handleFeeds(a *application.App, cmd application.Command) error {
	feeds, err := a.DB.GetFeeds(context.Background())
	if err != nil {
		return err
	}
	for _, feed := range feeds {
		usr, err := a.DB.GetUserByID(context.Background(), feed.UserID)
		if err != nil {
			return err
		}
		fmt.Println("* ", feed.Name)
		fmt.Println("* ", feed.Url)
		fmt.Println("* ", usr.Name)
	}

	return nil
}

func handleAddFeed(a *application.App, cmd application.Command) error {
	nbrArgs := 2
	err := checkCMDArgs(cmd, nbrArgs)
	if err != nil {
		return err
	}
	currentUsername, err := a.Config.GetCurrentUser()
	if err != nil {
		return err
	}
	user, err := a.DB.GetUserByName(context.Background(), currentUsername)
	if err != nil {
		return err
	}

	// feed, err := getfeed(cmd.Arguments[0], cmd.Arguments[1])
	// if err != nil {
	// 	return err
	// }
	createFeedParams := database.CreateFeedParams{
		ID:        int32(uuid.New().ID()),
		Name:      cmd.Arguments[0],
		Url:       cmd.Arguments[1],
		UserID:    user.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	createdFeed, err := a.DB.CreateFeed(context.Background(), createFeedParams)
	if err != nil {
		return err
	}

	createdFeedFollowParam := database.CreateFeedFollowParams{
		ID:        int32(uuid.New().ID()),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedsID:   createdFeed.ID,
	}
	_, err = a.DB.CreateFeedFollow(context.Background(), createdFeedFollowParam)
	if err != nil {
		fmt.Println("error creating associated feed_follow")
		return err
	}

	fmt.Printf("%+v\n", createdFeed)

	return nil
}

// func getfeed(url string) (*RSSFeed, error) {
// 	ctx, cancelFunc := context.WithTimeout(context.Background(), 3*time.Second)
// 	defer cancelFunc()
// 	rssFeed, err := fetchFeed(ctx, url)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return rssFeed, nil
// }

func handleAgg(a *application.App, cmd application.Command) error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelFunc()
	rssFeed, err := fetchFeed(ctx, "https://www.wagslane.dev/index.xml")
	if err != nil {
		return err
	}
	fmt.Printf("Fetched Feed: %+v\n", rssFeed)
	return nil
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {

	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "gator")

	client := http.DefaultClient

	res, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	resData, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var rssFeed RSSFeed

	err = xml.Unmarshal(resData, &rssFeed)
	if err != nil {
		return nil, err
	}

	rssFeed.Channel.Title = html.UnescapeString(rssFeed.Channel.Title)
	rssFeed.Channel.Description = html.UnescapeString(rssFeed.Channel.Description)

	for _, item := range rssFeed.Channel.Item {
		item.Title = html.UnescapeString(item.Title)
		item.Description = html.UnescapeString(item.Description)
	}

	return &rssFeed, nil
}

func handleLogin(a *application.App, cmd application.Command) error {
	err := checkCMDArgs(cmd)
	if err != nil {
		return err
	}

	userName := cmd.Arguments[0]

	user, err := a.DB.GetUserByName(context.Background(), userName)
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
	err := checkCMDArgs(cmd)
	if err != nil {
		return err
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
	fmt.Printf("New user registered: %+v\n", newUser)

	return nil
}

func handleReset(a *application.App, cmd application.Command) error {
	err := a.DB.DeleteAllUsers(context.Background())
	if err != nil {
		return err
	}
	fmt.Println("users table emptied!")
	return nil
}

func handleUsers(a *application.App, cmd application.Command) error {
	users, err := a.DB.GetUsers(context.Background())
	if err != nil {
		return err
	}

	currentUsername, err := a.Config.GetCurrentUser()
	if err != nil {
		return err
	}
	for _, u := range users {
		if u.Name == currentUsername {
			fmt.Printf("* %s (current)\n", u.Name)
		} else {
			fmt.Printf("* %s\n", u.Name)
		}
	}

	return nil
}

func checkCMDArgs(cmd application.Command, nbrArgs ...int) error {
	var maxArgs int
	if len(nbrArgs) > 0 {
		maxArgs = nbrArgs[0]
		if len(cmd.Arguments) < maxArgs {
			return fmt.Errorf(" %s command expects %d argument(s)", cmd.Name, maxArgs)
		}
		for _, arg := range cmd.Arguments[:maxArgs] {
			if len(strings.Trim(arg, " ")) < 1 {
				return fmt.Errorf("invalid argument provided")
			}
		}
	} else {
		if len(cmd.Arguments) < 1 {
			return fmt.Errorf(" %s command expects one or more argument(s)", cmd.Name)
		}
	}
	if len(strings.Trim(cmd.Arguments[0], " ")) < 1 {
		return fmt.Errorf("invalid argument provided")
	}
	return nil
}
