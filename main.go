package main

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/davidsbond/sse-cluster/broker"
	"github.com/davidsbond/sse-cluster/handler"
	"github.com/davidsbond/sse-cluster/config"
	"github.com/hashicorp/memberlist"
	"github.com/rs/xid"
)

func main() {
	cnf, err := config.Load()

	if err != nil {
		logrus.WithError(err).Fatal("failed to load configuration")
	}

	list, node, err := createMemberList(cnf)

	if err != nil {
		logrus.WithError(err).Fatal("failed to create member list")
	}

	br := broker.New(list, node)
	hnd := handler.New(br)
	mux := mux.NewRouter()

	mux.HandleFunc("/subscribe/{channel}", hnd.Subscribe).Methods("GET")
	mux.HandleFunc("/publish", hnd.Publish).Methods("POST")
	http.ListenAndServe(":"+cnf.Port, mux)
}

func createMemberList(cnf *config.Config) (*memberlist.Memberlist, *memberlist.Node, error) {
	logrus.SetOutput(os.Stdout)

	c := memberlist.DefaultLANConfig()

	if cnf.GossipPort > 0 {
		c.BindPort = cnf.GossipPort
	} else {
		c.BindPort = 0
	}

	c.Name = xid.New().String()

	list, err := memberlist.Create(c)

	if err != nil {
		return nil, nil, err
	}

	node := list.LocalNode()
	node.Meta = []byte(cnf.Port)

	if len(cnf.GossipAddrs) > 0 {
		if _, err := list.Join(cnf.GossipAddrs); err != nil {
			return nil, nil, err
		}
	}

	return list, node, nil
}
