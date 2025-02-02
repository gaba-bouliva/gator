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
	"strconv"
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
	app.RegisterCMD("users", middlewareLoggedIn(handleUsers))
	app.RegisterCMD("agg", middlewareLoggedIn(handleAgg))
	app.RegisterCMD("addfeed", middlewareLoggedIn(handleAddFeed))
	app.RegisterCMD("feeds", middlewareLoggedIn(handleFeeds))
	app.RegisterCMD("follow", middlewareLoggedIn(handleFollow))
	app.RegisterCMD("following", middlewareLoggedIn(handleFollowing))
	app.RegisterCMD("unfollow", middlewareLoggedIn(unfollow))
	app.RegisterCMD("browse", middlewareLoggedIn(handleBrowse))

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

func scrapeFeeds(a *application.App) error {
	nextFeed, err := a.DB.GetNextFeedToFetch(context.Background())
	if err != nil {
		return err
	}
	markFeedFetchedParams := database.MarkFeedFetchedParams{
		LastFetchedAt: sql.NullTime{Time: time.Now(), Valid: true},
		UpdatedAt:     time.Now(),
		ID:            nextFeed.ID,
	}
	err = a.DB.MarkFeedFetched(context.Background(), markFeedFetchedParams)
	if err != nil {
		return err
	}
	rssFeed, err := fetchFeed(context.Background(), nextFeed.Url)
	if err != nil {
		return err
	}

	for _, item := range rssFeed.Channel.Items {
		fmt.Println(item.Title)
		_, err := a.DB.GetPostByUrl(context.Background(), item.Link)
		if err == nil {
			continue
		}
		pubDate, err := tryParseDate(item.PubDate)
		if err != nil {
			return err
		}
		createdPostParams := database.CreatePostParams{
			ID:          int32(uuid.New().ID()),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			PublishedAt: pubDate,
			Title:       item.Title,
			Description: item.Description,
			Url:         item.Link,
			FeedID:      nextFeed.ID,
		}
		_, err = a.DB.CreatePost(context.Background(), createdPostParams)
		if err != nil {
			return err
		}
	}

	return nil
}

func handleBrowse(a *application.App, cmd application.Command, user database.User) error {
	limit := 2
	if len(cmd.Arguments) > 0 {
		parseIntArg, err := strconv.Atoi(cmd.Arguments[0])
		if err != nil {
			return err
		}
		limit = parseIntArg
	}
	posts, err := a.DB.GetPosts(context.Background(), int32(limit))
	if err != nil {
		return err
	}
	for _, post := range posts {
		fmt.Printf("* %+v", post)
		fmt.Println()
	}
	return nil
}

func tryParseDate(dateStr string) (time.Time, error) {
	formats := []string{
		time.RFC1123,
		time.RFC1123Z,
		time.RFC3339,
		time.RFC3339Nano,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		"2006-01-02 15:04:05",  // Example: 2025-02-02 14:30:00
		"2006-01-02T15:04:05Z", // Example: 2025-02-02T14:30:00Z
		"02 Jan 2006",          // Example: 02 Feb 2025
		"2006-01-02",           // Example: 2025-02-02
		"02/01/2006",           // Example: 02/02/2025
	}

	for _, format := range formats {
		t, err := time.Parse(format, dateStr)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

func handleFollowing(a *application.App, cmd application.Command, user database.User) error {

	usrFeedFollowings, err := a.DB.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return err
	}

	for _, feedsFollow := range usrFeedFollowings {
		fmt.Println("feed name: ", feedsFollow.FeedName)
	}

	return nil
}

func handleFollow(a *application.App, cmd application.Command, user database.User) error {
	nbrArgs := 1
	err := checkCMDArgs(cmd, nbrArgs)
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

func unfollow(a *application.App, cmd application.Command, user database.User) error {

	err := checkCMDArgs(cmd, 1)
	if err != nil {
		return err
	}

	url := cmd.Arguments[0]
	feed, err := a.DB.GetFeedByURL(context.Background(), url)
	if err != nil {
		return err
	}

	deleteFeedFollowParam := database.DeleteFeedFollowParams{
		UserID:  user.ID,
		FeedsID: feed.ID,
	}

	err = a.DB.DeleteFeedFollow(context.Background(), deleteFeedFollowParam)
	if err != nil {
		return err
	}

	return nil

}

func handleFeeds(a *application.App, cmd application.Command, user database.User) error {
	feeds, err := a.DB.GetFeeds(context.Background())
	if err != nil {
		return err
	}
	for _, feed := range feeds {
		fmt.Println("* ", feed.Name)
		fmt.Println("* ", feed.Url)
		fmt.Println("* ", user.Name)
	}

	return nil
}

func handleAddFeed(a *application.App, cmd application.Command, user database.User) error {
	nbrArgs := 2
	err := checkCMDArgs(cmd, nbrArgs)
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

func handleAgg(a *application.App, cmd application.Command, user database.User) error {
	err := checkCMDArgs(cmd, 1)
	if err != nil {
		return err
	}
	reqWaitTime, err := time.ParseDuration(cmd.Arguments[0])
	if err != nil {
		return err
	}
	fmt.Println("Collecting feeds every ", reqWaitTime)
	ticker := time.NewTicker(reqWaitTime)
	defer ticker.Stop()
	for range ticker.C {
		err := scrapeFeeds(a)
		if err != nil {
			return err
		}
	}

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

	for _, item := range rssFeed.Channel.Items {
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

	usr, err := a.DB.GetUserByName(context.Background(), userName)
	if err != nil {
		return err
	}

	err = a.Config.SetUser(usr.Name)
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

func handleUsers(a *application.App, cmd application.Command, user database.User) error {
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
