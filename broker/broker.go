// Package broker contains the broker implementataion.
package broker

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
	"sync"

	"github.com/hashicorp/memberlist"
	"github.com/sirupsen/logrus"
)

type (
	// The Broker type represents a node in the cluster, it contains
	// the list of all other members as well as a map of connected
	// client channels.
	Broker struct {
		memberlist Memberlist
		http       *http.Client
		mux        sync.Mutex
		channels   map[string]*Channel
		log        *logrus.Entry
		wg         sync.WaitGroup
	}

	// The Memberlist type represents the gossip implementation used by the
	// broker for service discovery.
	Memberlist interface {
		NumMembers() int
		LocalNode() *memberlist.Node
		Members() []*memberlist.Node
	}

	// The Status type represents the status of a node/cluster. It contains
	// sections for the gossip memberlist and the node's channels
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
func New(ml Memberlist, cl *http.Client) *Broker {
	br := &Broker{
		memberlist: ml,
		channels:   make(map[string]*Channel),
		http:       cl,
		log: logrus.WithFields(logrus.Fields{
			"name":     "broker",
			"brokerId": ml.LocalNode().Name,
		}),
	}

	return br
}

// Close blocks the goroutine until all asynchronous operations of the broker
// have stopped.
func (b *Broker) Close() {
	b.wg.Wait()
}

// Status returns information on the broker. It contains the number of running
// goroutines, the gossip members and total member count, as well as client information
// for this broker.
func (b *Broker) Status() *Status {
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

// Publish writes a given message to a client. If no client identifier is specified,
// the message is written to the entire channel. If running in a cluster, the event
// is forwarded asynchronously via HTTP to the next node whose id does not exist in
// the message's BeenTo field.
func (b *Broker) Publish(channelID, clientID string, msg Message) error {
	switch {
	// If we've got no channel or client identifier, publish to all clients
	// on all channels
	case channelID == "" && clientID == "":
		b.wg.Add(1)
		go b.publishAll(msg)
	// If we've got a channel identifier but no client identifier, write to
	// the entire channel
	case channelID != "" && clientID == "":
		b.wg.Add(1)
		go b.publishChannel(channelID, msg)
	// If we've got both a client and channel identifier, write to the client
	// on the given channel
	case channelID != "" && clientID != "":
		b.wg.Add(1)
		go b.publishClient(channelID, clientID, msg)
	default:
		return errors.New("invalid channel/client identifier combination")
	}

	// If we're not the only member, propagate the event
	if b.memberlist.NumMembers() > 1 {
		b.wg.Add(1)
		go b.sendToNextNode(channelID, clientID, msg)
	}

	return nil
}

func (b *Broker) publishAll(msg Message) {
	b.mux.Lock()
	defer b.mux.Unlock()
	defer b.wg.Done()

	for _, ch := range b.channels {
		b.mux.Unlock()

		ch.Write(msg)

		b.mux.Lock()
	}
}

func (b *Broker) publishChannel(channelID string, msg Message) {
	b.mux.Lock()
	defer b.mux.Unlock()
	defer b.wg.Done()

	if ch, ok := b.channels[channelID]; ok {
		ch.Write(msg)
	}
}

func (b *Broker) publishClient(channelID, clientID string, msg Message) {
	b.mux.Lock()
	defer b.mux.Unlock()
	defer b.wg.Done()

	if ch, ok := b.channels[channelID]; ok {
		ch.WriteTo(clientID, msg)
	}
}

func (b *Broker) sendToNextNode(channelID, clientID string, msg Message) {
	defer b.wg.Done()

	// Obtain the individual node ids from the X-Been-To header
	ids := make(map[string]interface{})
	for _, nodeID := range msg.BeenTo {
		ids[nodeID] = true
	}

	// For each member in the list
	for _, member := range b.memberlist.Members() {
		evtInfo := logrus.Fields{
			"targetNodeId": member.Name,
			"eventId":      msg.ID,
			"event":        msg.Event,
			"channel":      channelID,
		}

		// If we're looking at ourselves, or a node the message has already
		// been through, skip.
		if _, ok := ids[member.Name]; ok || member == b.memberlist.LocalNode() {
			continue
		}

		// Append this node's id to the list of node ids this event
		// has already been to
		msg.BeenTo = append(msg.BeenTo, b.memberlist.LocalNode().Name)
		url := fmt.Sprintf("http://%s:%s/publish/%s", member.Addr, member.Meta, channelID)

		// Send an HTTP POST request to the event publishing endpoint of the member
		// node.
		resp, err := b.http.Post(url, "application/json", bytes.NewBuffer(msg.JSON()))

		if err != nil {
			b.log.
				WithFields(evtInfo).
				WithError(err).
				Error("failed to perform http request")

			continue
		}

		// The publish endpoint should return a 200
		if resp.StatusCode != http.StatusOK {
			// If not, log the error and try the next node
			data, _ := ioutil.ReadAll(resp.Body)
			err := fmt.Errorf(string(data))

			b.log.
				WithFields(evtInfo).
				WithError(err).
				Error("failed to propagate event to node")

			continue
		}

		b.log.
			WithFields(evtInfo).
			Info("propagated message to node")

		// If we were successful, break, we will write the message to the first
		// node that isn't in the been to list
		break
	}
}

// NewClient creates a new client for a given channel. If the channel does not
// exist, it is created.
func (b *Broker) NewClient(channelID, clientID string) (*Client, error) {
	b.mux.Lock()
	defer b.mux.Unlock()

	ch, ok := b.channels[channelID]

	if !ok {
		ch = NewChannel(channelID)
		b.channels[channelID] = ch

		b.log.WithFields(logrus.Fields{
			"channel": channelID,
		}).Info("created new channel")
	}

	b.log.WithFields(logrus.Fields{
		"channel": channelID,
		"client":  clientID,
	}).Info("creating new client")

	return ch.AddClient(clientID)
}

// RemoveClient removes a client from a channel. If the channel has no
// connected clients, it is also removed.
func (b *Broker) RemoveClient(channelID, clientID string) {
	b.mux.Lock()
	defer b.mux.Unlock()

	channel, ok := b.channels[channelID]

	if !ok {
		return
	}

	channel.RemoveClient(clientID)

	b.log.WithFields(logrus.Fields{
		"channel": channelID,
		"client":  clientID,
	}).Info("removed client from channel")

	if channel.NumClients() == 0 {
		delete(b.channels, channelID)

		b.log.WithFields(logrus.Fields{
			"channel": channelID,
		}).Info("removed empty channel")
	}
}
