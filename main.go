package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/urfave/cli/v2"
)

const (
	ARGS_MODE            string = "mode"
	ARGS_MODE_SHORT      string = "m"
	ARGS_MODE_DES        string = "Run build/run/test"
	ARGS_ROOT_PATH       string = "root-path"
	ARGS_ROOT_PATH_SHORT string = "r"
	ARGS_ROOT_PATH_DES   string = "Set root-path to watch"
	ARGS_PATTERNS        string = "patterns"
	ARGS_PATTERNS_SHORT  string = "n"
	ARGS_PATTERNS_DES    string = "Specify file patterns"

	ERR_FMT_MSG            string = "Error: %v"
	ERR_INVALID_BUILD_MODE string = "Invalid build mode: %s options = build|test"

	MSG_BUILD_START string = "Build Started..."

	MSG_SUCCESS string = "SUCCESS"
	MSG_FAILED  string = "FAILED"

	MODE_BUILD string = "build"
	MODE_RUN   string = "run"
	MODE_TEST  string = "test"
)

type Options struct {
	BuildMode string
	Filepath  string
	Patterns  []string
}

func check_mode(mode string) bool {
	return mode != MODE_BUILD ||
		// mode != MODE_RUN ||
		mode != MODE_TEST
}

func go_build(o *Options) {
	log.Println(MSG_BUILD_START)
	cmd := exec.Command("go", "build")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = o.Filepath
	err := cmd.Run()
	if err != nil {
		log.Printf(ERR_FMT_MSG, err)
		log.Println(MSG_FAILED)
	} else {
		log.Println(MSG_SUCCESS)
	}
}

func go_run(o *Options) {
	cmd := exec.Command("go", "run", ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = o.Filepath
	err := cmd.Run()
	if err != nil {
		log.Printf(ERR_FMT_MSG, err)
	}
}

func go_test(o *Options) {
	cmd := exec.Command("go", "test", "-v")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = o.Filepath
	err := cmd.Run()
	if err != nil {
		log.Printf(ERR_FMT_MSG, err)
	}
}

func handle_event(o *Options, event fsnotify.Event) {
	for _, pattern := range o.Patterns {
		match, err := filepath.Match(pattern, event.Name)
		if err != nil {
			log.Printf(ERR_FMT_MSG, err)
		}

		if event.Op == fsnotify.Write ||
			event.Op == fsnotify.Create ||
			event.Op == fsnotify.Remove {
			if match {
				switch o.BuildMode {
				case MODE_BUILD:
					go_build(o)
				case MODE_RUN:
					go_run(o)
				case MODE_TEST:
					go_test(o)
				}
			}
		}
	}
}

func handle_error(o *Options, err error) {
	log.Printf(ERR_FMT_MSG, err)
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
				handle_event(o, event)
			case err := <-watcher.Errors:
				handle_error(o, err)
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
			Name:    ARGS_MODE,
			Aliases: []string{ARGS_MODE_SHORT},
			Usage:   ARGS_PATTERNS_DES,
		},
		&cli.StringFlag{
			Name:    ARGS_ROOT_PATH,
			Aliases: []string{ARGS_ROOT_PATH_SHORT},
			Usage:   ARGS_ROOT_PATH_DES,
		},
		&cli.StringSliceFlag{
			Name:    ARGS_PATTERNS,
			Aliases: []string{ARGS_PATTERNS_SHORT},
			Usage:   ARGS_PATTERNS_DES,
		},
	}

	app.Action = func(c *cli.Context) error {
		options := Options{
			BuildMode: MODE_BUILD,
			Filepath:  ".",
			Patterns:  []string{"*.go"},
		}

		arg_mode := c.String(ARGS_MODE)
		arg_filepath := c.String(ARGS_ROOT_PATH)
		arg_filepatterns := c.StringSlice(ARGS_PATTERNS)

		if arg_mode != "" && check_mode(arg_mode) {
			options.BuildMode = arg_mode
		} else if !check_mode(arg_mode) {
			log.Fatalf(ERR_INVALID_BUILD_MODE, arg_mode)
		}

		if arg_filepath != "" {
			options.Filepath = arg_filepath
		}

		if len(arg_filepatterns) != 0 {
			joined := append(options.Patterns, arg_filepatterns...)
			options.Patterns = Unique(joined)
		}

		watch(&options)
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
