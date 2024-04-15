package main

import (
	"fmt"
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/urfave/cli/v2"
)

const (
	ARGS_ROOT_PATH       string = "root-path"
	ARGS_ROOT_PATH_SHORT string = "r"
	ARGS_PATTERNS        string = "patterns"
	ARGS_PATTERNS_SHORT  string = "n"
)

type Options struct {
	Filepath  string
	Filenames []string
}

func watch(o *Options) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)

	// Start a go routine to watch for events
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				log.Println("Event:", event)
			case err := <-watcher.Errors:
				log.Println("Error:", err)
			}
		}
	}()
	err = watcher.Add(o.Filepath)
	if err != nil {
		log.Fatal(err)
	}

	<-done
}

func main() {
	app := cli.NewApp()
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    ARGS_ROOT_PATH,
			Aliases: []string{ARGS_ROOT_PATH_SHORT},
			Usage:   "Set root-path to watch",
		},
		&cli.StringSliceFlag{
			Name:    ARGS_PATTERNS,
			Aliases: []string{ARGS_PATTERNS_SHORT},
			Usage:   "Specify file patterns",
		},
	}

	app.Action = func(c *cli.Context) error {
		options := Options{
			Filepath:  ".",
			Filenames: []string{"*.go"},
		}

		if c.String(ARGS_ROOT_PATH) != "" {
			options.Filepath = ARGS_ROOT_PATH
		}

		if len(c.StringSlice(ARGS_PATTERNS)) != 0 {
			options.Filenames = c.StringSlice(ARGS_PATTERNS)
		}

		fmt.Println("Config:", options.Filepath)
		fmt.Println("Patterns:", options.Filenames)

		watch(&options)

		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
