package main

import (
	"fmt"
  "log"
  "os"

	"github.com/urfave/cli/v2"
  "gopkg.in/yaml.v2"
)

type Config struct {
  DefaultService string `yaml:"default_service"`
  Commands []string `yaml:"commands"`
}

func main() {
	app := &cli.App{
    Name: "Rudder - Make easier to use docker-compose.",
    Usage: "rudder CMD",
    Action: func(c *cli.Context) error {
      fmt.Println("boom! I say!")
      return nil
    },
  }

  err := app.Run(os.Args)
  if err != nil {
    log.Fatal(err)
  }
}
