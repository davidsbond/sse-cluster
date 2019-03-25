// Package main contains the entrypoint of the application
package main

import (
	"os"

	"github.com/davidsbond/sse-cluster/cmd"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	version string
)

func main() {
	logrus.SetOutput(os.Stdout)
	logrus.SetFormatter(&logrus.JSONFormatter{})

	app := cli.NewApp()
	app.Name = "sse-cluster"
	app.Author = "David Bond"
	app.Email = "davidsbond93@gmail.com"
	app.Version = version
	app.Usage = "Commands for running an SSE broker node"

	app.Commands = []cli.Command{
		cmd.Start(),
	}

	app.Run(os.Args)
}
