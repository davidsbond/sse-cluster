# sse-cluster

A scalable [Server Sent Events](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events) broker

## Features

* Channels
  * Seperate streams of events, they are created dynamically when the first client subscribes and are deleted automatically when the last client disconnects.
* Scalability
  * Each node uses [gossip protocol](https://en.wikipedia.org/wiki/Gossip_protocol) to discover more nodes. New nodes need only be started with the hostname of a single active node in the cluster.
  * When a node recieves an event, it propagates it to the next node, appending metadata to the message to avoid event duplication
* `EventSource` compatibility

  * Using JavaScript, you can use native `EventSource` class to stream events from the broker. Below is an example:

```javascript
  const channel = 'my-channel'

  // Connect to a single node, or to a load balancer in-front
  // of many nodes
  const es = new EventSource(`https://my-sse-cluster:8080/subscribe/${channel}`)

  // Handle the stream generically, all events will trigger this method
  es.onmessage = (e) => {
    const newElement = document.createElement('li');
    const eventList = document.getElementById('list');

    newElement.innerHTML = `message: ${e.data}`;
    eventList.appendChild(newElement);
  }

  // Handle specific event types, this method is invoked for 'ping' events.
  es.addEventListener('ping', function(e) {
    const newElement = document.createElement('li');
    const obj = JSON.parse(e.data);

    newElement.innerHTML = `ping at ${obj.time}`;
    eventList.appendChild(newElement);
  }, false);
```

## Installation

Each node can be ran as a single binary, docker image or Kubernetes deployment.

### Installing from source

This section assumes you have [go 1.11+](https://golang.org) installed.

```bash
# download the source code
go get github.com/davidsbond/sse-cluster

# install the binary
go install github.com/davidsbond/sse-cluster

# start a node
sse-cluster start
```

### Running with docker

The application is also available as a docker image [here](https://hub.docker.com/_/golang/)

```bash
docker run -d davidsbond/sse-cluster start
```

### Installing with helm

This repository also contains a [helm chart](https://helm.sh) for deploying to [Kubernetes](https://kubernetes.io) clusters.

```bash
helm install my-sse-cluster ./helm/sse-cluster/
```

Upon success, a `StatefulSet` and `HorizontalPodAutoscaler` will be created in your Kubernetes cluster that will manage the number of sse-brokering nodes. A [headless service](https://kubernetes.io/docs/concepts/services-networking/service/#headless-services) will also be created to allow name resolution of individual pods. Without this, the sse broker nodes will not be able to communicate.

## Configuration

Configuration is provided to a node either via environment variables or command-line arguments:

| Argument       | Environment Variable | Description                                                                               | Default       |
|:---------------|:---------------------|:------------------------------------------------------------------------------------------|:--------------|
| `gossip.port`  | `GOSSIP_PORT`        | The port to use for communications via gossip protocol                                    | `<generated>` |
| `gossip.hosts` | `GOSSIP_HOSTS`       | The initial hosts the node should connect to, should be a comma-seperated string of hosts | `N/A`         |
| `http.port`    | `HTTP_PORT`          | The port to use for HTTP communication                                                    | `8080`        |