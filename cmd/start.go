package cmd

import (
	"context"
	"log"
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
				Usage:  "The initial hosts the node should connect to, should be a comma-seperated string of hosts",
			},
			cli.IntFlag{
				Usage:  "The port to use for communications via gossip protocol",
				Name:   "gossip.port",
				EnvVar: "GOSSIP_PORT",
				Value:  42000,
			},
			cli.StringFlag{
				Usage:  "The key used to initialize the primary encryption key in a keyring",
				Name:   "gossip.secretKey",
				EnvVar: "GOSSIP_SECRET_KEY",
			},
			cli.StringFlag{
				Usage:  "The port to use for listening to HTTP requests",
				Name:   "http.server.port",
				EnvVar: "HTTP_SERVER_PORT",
				Value:  "8080",
			},
			cli.DurationFlag{
				Name:   "http.client.timeout",
				Usage:  "Sets the request timeout for the http client",
				EnvVar: "HTTP_CLIENT_TIMEOUT",
				Value:  time.Second * 10,
			},
		},
	}
}

func start(ctx *cli.Context) error {
	list, err := createMemberList(ctx)

	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	cl := &http.Client{
		Timeout: ctx.Duration("http.client.timeout"),
	}

	br := broker.New(list, cl)
	hnd := handler.New(br)
	svr := createHTTPServer(ctx, hnd)

	// Execute ListenAndServe in a seperate goroutine as it blocks
	go func() {
		logrus.Info("starting http server")

		if err := svr.ListenAndServe(); err != nil {
			logrus.WithError(err).Error("http server exited")
		}
	}()

	if err := handleExitSignal(br, svr, list); err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	return nil
}

func handleExitSignal(b *broker.Broker, svr *http.Server, ml *memberlist.Memberlist) error {
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
	if err := svr.Shutdown(ctx); err != nil {
		return err
	}

	// Wait for any broker operations to finish
	b.Close()

	return logrus.StandardLogger().Writer().Close()
}

func createHTTPServer(ctx *cli.Context, h *handler.Handler) *http.Server {
	mux := mux.NewRouter()

	mux.HandleFunc("/status", h.Status).Methods("GET")
	mux.HandleFunc("/subscribe/{channel}", h.Subscribe).Methods("GET")
	mux.HandleFunc("/publish/{channel}", h.Publish).
		Methods("POST").
		Headers("Content-Type", "application/json")

	mux.HandleFunc("/publish/{channel}/client/{client}", h.Publish).
		Methods("POST").
		Headers("Content-Type", "application/json")

	svr := &http.Server{
		Handler:  mux,
		Addr:     ":" + ctx.String("http.server.port"),
		ErrorLog: log.New(logrus.StandardLogger().Writer(), "", 0),
	}

	return svr
}

func createMemberList(ctx *cli.Context) (*memberlist.Memberlist, error) {
	c := memberlist.DefaultLANConfig()

	c.Logger = log.New(logrus.StandardLogger().Writer(), "", 0)
	c.BindPort = ctx.Int("gossip.port")
	c.SecretKey = []byte(ctx.String("gossip.secret-key"))

	logrus.Info("creating gossip memberlist")

	list, err := memberlist.Create(c)

	if err != nil {
		return nil, err
	}

	hosts := ctx.StringSlice("gossip.hosts")
	hostname, _ := os.Hostname()

	actual := []string{}
	for _, host := range hosts {
		if strings.Contains(host, hostname) {
			continue
		}

		actual = append(actual, host)
	}

	list.LocalNode().Meta = []byte(ctx.String("http.server.port"))

	if len(actual) > 0 {
		logrus.WithField("hosts", actual).Info("joining sse cluster")

		if _, err := list.Join(actual); err != nil {
			return nil, err
		}
	}

	return list, nil
}
