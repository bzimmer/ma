package main

//go:generate go run main.go manual -o ../../docs/commands.md ../../docs/commands

import (
	"context"
	"fmt"
	"os"

	"github.com/bzimmer/ma/internal"
	"github.com/rs/zerolog/log"
)

func main() {
	app := internal.App()
	var err error
	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case error:
				err = v
			default:
				err = fmt.Errorf("%v", v)
			}
		}
		if err != nil {
			log.Error().Err(err).Msg(app.Name)
			os.Exit(1)
		}
		os.Exit(0)
	}()
	err = app.RunContext(context.Background(), os.Args)
}
