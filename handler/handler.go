// Package handler contains types for interacting with the broker implementation
// via HTTP.
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/davidsbond/sse-cluster/broker"
	"github.com/gorilla/mux"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
)

type (
	// The Handler type contains methods for handling inbound HTTP requests
	// to the broker.
	Handler struct {
		broker Broker
		log    *logrus.Entry
	}

	// The Broker interface defines methods the HTTP handlers use to perform
	// operations against the broker from HTTP requests.
	Broker interface {
		Status() *broker.Status
		Publish(string, string, broker.Message) error
		NewClient(string, string) (*broker.Client, error)
		RemoveClient(string, string)
	}
)

// New creates a new instance of the Handler type with the given broker
func New(br Broker) *Handler {
	return &Handler{
		broker: br,
		log:    logrus.WithField("name", "handler"),
	}
}

// Status handles an incoming HTTP GET request that returns the current
// status of the node and the gossip member list
func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	status := h.broker.Status()

	if err := json.NewEncoder(w).Encode(status); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
}

// Publish handles an incoming HTTP POST request and writes a message to the broker.
// Returns a 400 if invalid JSON has been provided.
func (h *Handler) Publish(w http.ResponseWriter, r *http.Request) {
	var msg broker.Message

	vars := mux.Vars(r)
	channelID := vars["channel"]
	clientID := vars["client"]

	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.broker.Publish(channelID, clientID, msg); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Subscribe handles an incoming HTTP GET request and starts an event-stream with
// the client. The connection remains open while events are read from the broker.
// Events are written sequentially in 'text/event-stream' format. When the client
// disconnects, they're removed from the broker.
func (h *Handler) Subscribe(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)

	if !ok {
		http.Error(w, "client does not support streaming", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)

	// Get the channel/client IDs from the url params
	channelID := vars["channel"]
	clientID, ok := vars["client"]

	if !ok {
		clientID = xid.New().String()
	}

	reqInfo := logrus.Fields{
		"client":  clientID,
		"channel": channelID,
		"host":    r.Host,
	}

	h.log.WithFields(reqInfo).Info("new subscriber connection")

	// Set streaming headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	client, err := h.broker.NewClient(channelID, clientID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for {
		select {
		case msg := <-client.Messages():
			if _, err := w.Write(msg.Bytes()); err != nil {
				h.log.WithError(err).WithFields(reqInfo).Error("failed to write data")
				continue
			}

			flusher.Flush()
		case <-r.Context().Done():
			h.broker.RemoveClient(channelID, clientID)
			h.log.WithFields(reqInfo).Info("subscriber disconnected")

			return
		}
	}
}
