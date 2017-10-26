package main

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
)

var HttpPorts map[string]string
var TlsPorts map[string]string

type Docker struct {
	ctx         context.Context
	cli         *client.Client
	networkName string
	networkID   string
	stopchan    chan int
}

func (x *Docker) AddNetwork(name string, containerID string, labels map[string]string) {
	if _, ok := labels["proxy"]; !ok {
		return
	}

	serviceName := labels["com.docker.compose.service"]
	stackName := labels["com.docker.compose.project"]
	host := stackName + "_" + serviceName
	aliases := []string{host, serviceName} //FIXME: add serviceName only for service mode
	log.Println("Adding container", name, "to network", x.networkName, aliases)

	options := network.EndpointSettings{}
	options.Aliases = aliases
	err := x.cli.NetworkConnect(x.ctx, x.networkID, containerID, &options)
	if err != nil {
		log.Errorln("Fail to add container", name, "to network", x.networkName, ":", err)
		return
	}

	if proxyHttpPort, ok := labels["proxy-http-port"]; ok {
		HttpPorts[host] = proxyHttpPort
		HttpPorts[serviceName] = proxyHttpPort //FIXME: add serviceName only for service mode
		log.Println("Map http port for service", host, "/", serviceName, "to", proxyHttpPort)
	} else {
		delete(HttpPorts, host)
		delete(HttpPorts, serviceName) //FIXME: add serviceName only for service mode
	}
	if proxyTlsPort, ok := labels["proxy-tls-port"]; ok {
		TlsPorts[host] = proxyTlsPort
		TlsPorts[serviceName] = proxyTlsPort //FIXME: add serviceName only for service mode
		log.Println("Map tls port for service", host, "/", serviceName, "to", proxyTlsPort)
	} else {
		delete(TlsPorts, host)
		delete(TlsPorts, serviceName) //FIXME: add serviceName only for service mode
	}
}

func (x *Docker) RemoveNetwork(name string, containerID string) {
	log.Println("Removing container", name, "from network", x.networkName)

	err := x.cli.NetworkDisconnect(x.ctx, x.networkID, containerID, true)
	if err != nil {
		log.Errorln("Fail to remove container", name, " from network", x.networkName, ":", err)
	}
}

func (x *Docker) Events(msgchan <-chan events.Message, errchan <-chan error) {
	log.Println("Listening to docker events...")
	for {
		select {
		case msg := <-msgchan:
			if msg.Type == events.ContainerEventType {
				if msg.Action == "start" {
					log.Println("Received", msg)
					if _, ok := msg.Actor.Attributes["proxy"]; ok {
						x.AddNetwork(msg.Actor.Attributes["name"], msg.Actor.ID, msg.Actor.Attributes)
					}
				} else if msg.Action == "die" {
					if _, ok := msg.Actor.Attributes["proxy"]; ok {
						x.RemoveNetwork(msg.Actor.Attributes["name"], msg.Actor.ID)
					}
				}
			}
		case err := <-errchan:
			log.Fatalln("err", err)
		case <-x.stopchan:
			log.Println("Stop listening to docker events...")
			break
		}
	}
}

func DockerInit(_publicNetworkName string) *Docker {
	var x Docker
	HttpPorts = make(map[string]string)
	TlsPorts = make(map[string]string)
	x.ctx = context.Background()
	x.networkName = _publicNetworkName
	var err error
	x.cli, err = client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	networkListOptions := types.NetworkListOptions{}
	networks, err := x.cli.NetworkList(x.ctx, networkListOptions)
	if err != nil {
		panic(err)
	}

	x.networkID = ""
	for _, network := range networks {
		if network.Name == x.networkName {
			x.networkID = network.ID
		}
	}
	if x.networkID == "" {
		log.Fatalln("Public Network for docker not found :", x.networkName)
	}

	eventOptions := types.EventsOptions{}
	msgchan, errchan := x.cli.Events(x.ctx, eventOptions)

	containerListOptions := types.ContainerListOptions{}
	containers, _ := x.cli.ContainerList(x.ctx, containerListOptions)
	for _, container := range containers {
		log.Println(container.Names[0], container.ID, container.State, container.Labels)

		if container.State == "running" && container.Labels["proxy"] == "true" {
			x.RemoveNetwork(container.Names[0], container.ID)
			x.AddNetwork(container.Names[0], container.ID, container.Labels)
		}
	}

	go x.Events(msgchan, errchan)

	return &x
}
