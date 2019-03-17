package handler

import (
	"encoding/json"
	"net/http"

	"github.com/davidsbond/sse-cluster/broker"
	"github.com/davidsbond/sse-cluster/message"
	"github.com/gorilla/mux"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
)

type (
	// The Handler type contains methods for handling inbound HTTP requests
	// to the broker.
	Handler struct {
		broker *broker.Broker
		log    *logrus.Entry
	}
)

// New creates a new instance of the Handler type with the given broker
func New(br *broker.Broker) *Handler {
	return &Handler{
		broker: br,
		log:    logrus.WithField("name", "handler"),
	}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	health := h.broker.GetStatus()

	if err := json.NewEncoder(w).Encode(health); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
}

// Publish handles an incoming HTTP POST request and writes a message to the broker.
// Returns a 400 if invalid JSON has been provided.
func (h *Handler) Publish(w http.ResponseWriter, r *http.Request) {
	var msg message.Message

	vars := mux.Vars(r)
	channelID := vars["channel"]

	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "invalid json in request", http.StatusBadRequest)
	}

	h.broker.Publish(channelID, msg)
}

// Subscribe handles an incoming HTTP GET request and starts an event-stream with
// the client. The connection remains open while events are read from the broker.
// Events are written sequentially in 'text/event-stream' format. When the client
// disconnects, they're removed from the broker.
func (h *Handler) Subscribe(w http.ResponseWriter, r *http.Request) {
	closer, cOK := w.(http.CloseNotifier)
	flusher, fOK := w.(http.Flusher)

	if !cOK || !fOK {
		http.Error(w, "client does not support streaming", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)

	// Generate a random id for the client, obtain the
	// channel id from the url
	clientID := xid.New().String()
	channelID := vars["channel"]

	h.log.WithFields(logrus.Fields{
		"clientId":  clientID,
		"channelId": channelID,
		"host":      r.Host,
	}).Info("new subscriber connection")

	// Set streaming headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	channel, client := h.broker.NewClient(channelID, clientID)

	for {
		select {
		case data := <-client.Messages():
			w.Write(data)
			flusher.Flush()
		case <-closer.CloseNotify():
			channel.RemoveClient(clientID)

			logrus.WithFields(logrus.Fields{
				"clientId":  clientID,
				"channelId": channelID,
				"host":      r.Host,
			}).Info("subscriber disconnected")

			return
		}
	}
}
