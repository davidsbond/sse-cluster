package cmd

import (
	"net/http"
	"os"
	"strings"

	"github.com/davidsbond/sse-cluster/broker"
	"github.com/davidsbond/sse-cluster/handler"
	"github.com/gorilla/mux"
	"github.com/hashicorp/memberlist"
	"github.com/urfave/cli"
)

// Start generates the start command that is used to create an SSE node.
func Start() cli.Command {
	return cli.Command{
		Name:   "start",
		Action: start,
		Usage:  "Start an SSE node",
		Flags: []cli.Flag{
			cli.StringSliceFlag{
				Name:   "gossip.hosts",
				EnvVar: "GOSSIP_HOSTS",
				Usage:  "Comma separated list of hosts to sync with",
			},
			cli.IntFlag{
				Usage:  "The port to use for gossip protocol",
				Name:   "gossip.port",
				EnvVar: "GOSSIP_PORT",
				Value:  42000,
			},
			cli.StringFlag{
				Usage:  "The port to use for http protocol",
				Name:   "http.port",
				EnvVar: "HTTP_PORT",
				Value:  "8080",
			},
		},
	}
}

func start(ctx *cli.Context) error {
	list, node, err := createMemberList(ctx)

	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	httpPort := ctx.String("http.port")

	br := broker.New(list, node, httpPort)
	hnd := handler.New(br)

	mux := mux.NewRouter()

	mux.HandleFunc("/status", hnd.Status).Methods("GET")
	mux.HandleFunc("/subscribe/{channel}", hnd.Subscribe).Methods("GET")
	mux.HandleFunc("/publish/{channel}", hnd.Publish).Methods("POST")

	if err := http.ListenAndServe(":"+httpPort, mux); err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	return nil
}

func createMemberList(ctx *cli.Context) (*memberlist.Memberlist, *memberlist.Node, error) {
	c := memberlist.DefaultLANConfig()

	c.BindPort = ctx.Int("gossip.port")
	c.Logger = nil

	list, err := memberlist.Create(c)

	if err != nil {
		return nil, nil, err
	}

	node := list.LocalNode()
	hosts := ctx.StringSlice("gossip.hosts")
	hostname, _ := os.Hostname()

	var actual []string
	for _, host := range hosts {
		if strings.Contains(host, hostname) {
			continue
		}

		actual = append(actual, host)
	}

	if len(hosts) > 0 {
		if _, err := list.Join(actual); err != nil {
			return nil, nil, err
		}
	}

	return list, node, nil
}
