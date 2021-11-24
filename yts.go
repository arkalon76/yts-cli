package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/fatih/color"
	"github.com/gosuri/uitable"
	"github.com/hekmon/transmissionrpc/v2"
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
)

const (
	YTSBaseURL        = "https://yts.mx/api/v2"
	YTSListMoviesPath = "/list_movies.json"
)

var (
	Quality map[string]int = map[string]int{
		"720p":  1,
		"1080p": 2,
		"2160p": 3,
	}
	transmissionDebug bool
)

func main() {
	app := &cli.App{
		Name:  "yts",
		Usage: "Get latest yts listings",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				EnvVars: []string{"YTS_CONFIG"},
				Usage:   "Load configuration from `FILE`",
			},
		},
		Action: func(c *cli.Context) error {
			showLatestMovies()
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:    "download",
				Aliases: []string{"d"},
				Usage:   "Download torrent with ID `n`",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "debug",
						Aliases: []string{"d"},
						Usage:   "Will enable debug mode on transmission client",
						Value:   false,
					},
				},
				Action: func(c *cli.Context) error {
					transmissionDebug = c.Bool("debug")
					arg := c.Args().First()
					torrent, err := strconv.ParseInt(arg, 10, 64)
					if err != nil {
						return err
					}

					entries, err := fetchLatestMovies()
					if err != nil {
						return err
					}

					highScore := 0
					var topPick string
					movie := entries.Data.Movies[torrent]
					torrentlist := entries.Data.Movies[torrent].Torrents

					for _, t := range torrentlist {
						if Quality[t.Quality] > highScore {
							highScore = Quality[t.Quality]
							topPick = t.URL
						}
					}
					fmt.Printf("Will download movie %s (%d)\n", movie.Title, movie.Year)
					return downloadTorrent(topPick)
				},
			},
			{
				Name:    "list",
				Aliases: []string{"l"},
				Usage:   "Will latest 20 movies on yts",
				Action: func(c *cli.Context) error {
					showLatestMovies()
					return nil
				},
			},
		},
	}

	loadConfig("config.yaml")
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}

func loadConfig(path string) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("Cant find config file")
		} else {
			fmt.Println("Found config but cant read it")
		}
	}
}

func showLatestMovies() {
	entries, err := fetchLatestMovies()
	if err != nil {
		panic(err)
	}

	table := uitable.New()
	table.MaxColWidth = 120
	table.Wrap = false
	table.AddRow(color.GreenString("No:"), color.GreenString("Title:"), color.GreenString("Year:"), color.GreenString("Rating:"), color.GreenString("Uploaded:"), color.GreenString("Synopsis:"))
	table.AddRow(color.GreenString("---"), color.GreenString("------"), color.GreenString("-----"), color.GreenString("-------"), color.GreenString("---------"), color.GreenString("---------"))
	for i, m := range entries.Data.Movies {
		table.AddRow(i, m.Title, m.Year, m.Rating, m.DateUploaded, m.Synopsis)
	}
	fmt.Println(table)
}

func fetchLatestMovies() (*YtsEntry, error) {
	var entry YtsEntry

	c := http.DefaultClient
	res, err := c.Get(fmt.Sprintf("%s/%s?sort_by=date_added&quality=1080p", YTSBaseURL, YTSListMoviesPath))
	if err != nil {
		return nil, err
	}
	bodybytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bodybytes, &entry); err != nil {
		return nil, err
	}

	return &entry, nil
}

func downloadTorrent(url string) error {
	host := viper.GetString("transmission.host")
	user := viper.GetString("transmission.user")
	pass := viper.GetString("transmission.pass")
	path := viper.GetString("transmission.destinationPath")

	bt, err := transmissionrpc.New(host, user, pass, &transmissionrpc.AdvancedConfig{Port: 9091, Debug: transmissionDebug})
	if err != nil {
		return err
	}
	_, err = bt.TorrentAdd(context.TODO(), transmissionrpc.TorrentAddPayload{
		Filename:    &url,
		DownloadDir: &path,
	})
	if err != nil {
		return err
	}
	return nil
}

type YtsEntry struct {
	Status        string `json:"status"`
	StatusMessage string `json:"status_message"`
	Data          struct {
		MovieCount int `json:"movie_count"`
		Limit      int `json:"limit"`
		PageNumber int `json:"page_number"`
		Movies     []struct {
			ID                      int      `json:"id"`
			URL                     string   `json:"url"`
			ImdbCode                string   `json:"imdb_code"`
			Title                   string   `json:"title"`
			TitleEnglish            string   `json:"title_english"`
			TitleLong               string   `json:"title_long"`
			Slug                    string   `json:"slug"`
			Year                    int      `json:"year"`
			Rating                  float64  `json:"rating"`
			Runtime                 int      `json:"runtime"`
			Genres                  []string `json:"genres"`
			Summary                 string   `json:"summary"`
			DescriptionFull         string   `json:"description_full"`
			Synopsis                string   `json:"synopsis"`
			YtTrailerCode           string   `json:"yt_trailer_code"`
			Language                string   `json:"language"`
			MpaRating               string   `json:"mpa_rating"`
			BackgroundImage         string   `json:"background_image"`
			BackgroundImageOriginal string   `json:"background_image_original"`
			SmallCoverImage         string   `json:"small_cover_image"`
			MediumCoverImage        string   `json:"medium_cover_image"`
			LargeCoverImage         string   `json:"large_cover_image"`
			State                   string   `json:"state"`
			Torrents                []struct {
				URL              string `json:"url"`
				Hash             string `json:"hash"`
				Quality          string `json:"quality"`
				Type             string `json:"type"`
				Seeds            int    `json:"seeds"`
				Peers            int    `json:"peers"`
				Size             string `json:"size"`
				SizeBytes        int    `json:"size_bytes"`
				DateUploaded     string `json:"date_uploaded"`
				DateUploadedUnix int    `json:"date_uploaded_unix"`
			} `json:"torrents"`
			DateUploaded     string `json:"date_uploaded"`
			DateUploadedUnix int    `json:"date_uploaded_unix"`
		} `json:"movies"`
	} `json:"data"`
	Meta struct {
		ServerTime     int    `json:"server_time"`
		ServerTimezone string `json:"server_timezone"`
		APIVersion     int    `json:"api_version"`
		ExecutionTime  string `json:"execution_time"`
	} `json:"@meta"`
}
