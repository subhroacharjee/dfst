package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/subhroacharjee/dfst/internal/config"
	"github.com/subhroacharjee/dfst/internal/store"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() (err error) {
	ctx, cancel := context.WithCancel(context.Background())

	defer func() {
		if r := recover(); r != nil {
			cancel()
			err = fmt.Errorf("recovered from panic: %v", r)
		}
	}()

	var c *config.Config

	if len(os.Args) < 2 {
		c = config.GetDefaultConfig()
	} else {
		c, err = config.ParseConfig(os.Args[1])
		if err != nil {
			return err
		}

	}

	store := store.NewStore(*c.GetStoreOpts())

	err = store.Start(ctx)
	return err
}
