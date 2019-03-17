package broker

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime"
	"sync"

	"github.com/davidsbond/sse-cluster/channel"
	"github.com/davidsbond/sse-cluster/client"
	"github.com/davidsbond/sse-cluster/message"

	"github.com/hashicorp/memberlist"
	"github.com/sirupsen/logrus"
)

type (
	// The Broker type represents a node in the cluster, it contains
	// the list of all other members as well as a map of connected
	// client channels.
	Broker struct {
		memberlist *memberlist.Memberlist
		node       *memberlist.Node

		mux      sync.Mutex
		channels map[string]*channel.Channel
	}

	Status struct {
		Goroutines int `json:"num_goroutines"`
		Gossip     struct {
			MemberCount int            `json:"member_count"`
			Members     map[string]int `json:"members"`
		} `json:"gossip"`
		Channels map[string][]string `json:"channels"`
	}
)

// New creates a new instance of the Broker type using the given member list and
// node.
func New(ml *memberlist.Memberlist, node *memberlist.Node) *Broker {
	br := &Broker{
		memberlist: ml,
		channels:   make(map[string]*channel.Channel),
		node:       node,
	}

	return br
}

// GetStatus returns information on the broker. It contains the number of running
// goroutines, the gossip members and total member count, as well as client information
// for this broker.
func (b *Broker) GetStatus() *Status {
	health := &Status{}

	health.Goroutines = runtime.NumGoroutine()

	health.Gossip.MemberCount = b.memberlist.NumMembers()
	health.Gossip.Members = make(map[string]int)

	for _, member := range b.memberlist.Members() {
		health.Gossip.Members[member.Addr.String()] = int(member.Port)
	}

	health.Channels = make(map[string][]string)

	b.mux.Lock()
	defer b.mux.Unlock()

	for id, channel := range b.channels {
		health.Channels[id] = channel.ClientIDs()
	}

	return health
}

// Publish writes a given message to a client channel.
func (b *Broker) Publish(channelID string, msg message.Message) {
	b.mux.Lock()

	// Write the message to the channel
	if ch, ok := b.channels[channelID]; ok {
		ch.Write(msg.Bytes())
	}

	b.mux.Unlock()

	// Obtain the individual node ids from the X-Been-To header
	ids := make(map[string]interface{})
	for _, nodeID := range msg.BeenTo {
		ids[nodeID] = true
	}

	// For each member in the list
	for _, member := range b.memberlist.Members() {
		// If we're looking at ourselves, or a node the message has already
		// been through, skip.
		if _, ok := ids[member.Name]; ok || member == b.node {
			continue
		}

		// Append this node's id to the list of node ids this event
		// has already been to
		msg.BeenTo = append(msg.BeenTo, b.node.Name)

		// Otherwise, create a new request to the publish endpoint for the list
		// member. They store their HTTP port in the metadata
		url := fmt.Sprintf("http://%s:%s/publish/%s", member.Addr, member.Meta, channelID)

		if _, err := http.Post(url, "application/json", bytes.NewBuffer(msg.JSON())); err != nil {
			logrus.WithError(err).Error("failed to build http request")
			continue
		}

		logrus.WithFields(logrus.Fields{
			"nodeId":  member.Name,
			"eventId": msg.ID,
			"event":   msg.Event,
			"channel": channelID,
		}).Info("propagated message to node")

		// If we were successful, break, we will write the message to the first
		// node that isn't in the been to list
		break
	}
}

// NewClient creates a new client for a given channel. If the channel does not
// exist, it is created.
func (b *Broker) NewClient(channelID, clientID string) (*channel.Channel, *client.Client) {
	b.mux.Lock()
	defer b.mux.Unlock()

	ch, ok := b.channels[channelID]

	if !ok {
		ch = channel.New(channelID)
		b.channels[channelID] = ch
	}

	cl := ch.AddClient(clientID)

	return ch, cl
}
