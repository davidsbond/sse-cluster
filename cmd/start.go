package cmd

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/davidsbond/sse-cluster/broker"
	"github.com/davidsbond/sse-cluster/handler"
	"github.com/gorilla/mux"
	"github.com/hashicorp/memberlist"
	"github.com/sirupsen/logrus"
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
	list, err := createMemberList(ctx)

	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	httpPort := ctx.String("http.port")

	br := broker.New(list, httpPort)
	hnd := handler.New(br)
	svr := createHTTPServer(ctx, hnd)

	// Execute ListenAndServe in a seperate goroutine as it blocks
	go func() {
		logrus.Info("starting http server")
	
		if err := svr.ListenAndServe(); err != nil {
			logrus.WithError(err).Error("http server exited")
		}
	}()

	if err := handleExitSignal(svr, list); err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	return nil
}

func handleExitSignal(svr *http.Server, ml *memberlist.Memberlist) error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop
	logrus.Info("got shutdown signal")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Leave the gossip memberlist
	if err := ml.Leave(time.Second * 5); err != nil {
		return err
	}
	
	// Gracefully shut down the HTTP server
	return svr.Shutdown(ctx)
}

func createHTTPServer(ctx *cli.Context, h *handler.Handler) *http.Server {
	mux := mux.NewRouter()

	mux.HandleFunc("/status", h.Status).Methods("GET")
	mux.HandleFunc("/subscribe/{channel}", h.Subscribe).Methods("GET")
	mux.HandleFunc("/publish/{channel}", h.Publish).Methods("POST")

	svr := &http.Server{
		Handler: mux,
		Addr:    ":" + ctx.String("http.port"),
	}

	return svr
}

func createMemberList(ctx *cli.Context) (*memberlist.Memberlist,error) {
	c := memberlist.DefaultLANConfig()

	c.BindPort = ctx.Int("gossip.port")
	c.LogOutput = nil

	logrus.Info("creating gossip memberlist")

	list, err := memberlist.Create(c)

	if err != nil {
		return nil, err
	}

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
		logrus.WithField("hosts", actual).Info("joining sse cluster")

		if _, err := list.Join(actual); err != nil {
			return nil, err
		}
	}

	return list, nil
}
