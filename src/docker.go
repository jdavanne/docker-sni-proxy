package main

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/swarm"
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

func (x *Docker) AddPublicNetworkToService(name string, serviceID string, service *swarm.Service) {
	if service == nil {
		servicep, _, err := x.cli.ServiceInspectWithRaw(x.ctx, serviceID, types.ServiceInspectOptions{})
		if err != nil {
			log.Errorln("[service] Error fetching service specs", name, serviceID, err)
			return
		}
		service = &servicep
	}
	serviceSpec := service.Spec
	labels := serviceSpec.TaskTemplate.ContainerSpec.Labels
	//log.Println("[service] service spec", name, serviceID, serviceSpec)
	//log.Println("[service] service spec networks", name, serviceID, serviceSpec.TaskTemplate.Networks)

	if _, ok := labels["proxy"]; !ok {
		log.Println("[service] Service is not public, skipping...", name, serviceID)
		return
	}

	serviceName := labels["com.docker.swarm.service.name"]
	host := serviceName
	aliases := []string{host}
	x.MapInternalPorts(host, serviceName, labels)

	// Skip network if not defined
	if x.networkID == "" {
		return
	}

	for _, network := range serviceSpec.TaskTemplate.Networks {
		if network.Target == x.networkID {
			log.Println("[service] Network already added service, skipping...", name, serviceID)
			return
		}
	}

	network := swarm.NetworkAttachmentConfig{}
	options := types.ServiceUpdateOptions{}
	network.Aliases = aliases
	network.Target = x.networkID
	serviceSpec.TaskTemplate.Networks = append(serviceSpec.TaskTemplate.Networks, network)

	//log.Println("[network] service spec (new)", name, serviceID, serviceSpec)
	//log.Println("[network] service spec networks(new)", name, serviceID, serviceSpec.TaskTemplate.Networks)
	log.Println("[network] Adding service", name, "to network", x.networkName, aliases)
	_, err := x.cli.ServiceUpdate(x.ctx, serviceID, service.Version, serviceSpec, options)
	if err != nil {
		log.Errorln("[network] Fail to add service", name, "to network", x.networkName, ":", err)
		if err.Error() == "Error response from daemon: rpc error: code = Unknown desc = update out of sequence" {
			log.Warn("[network] Retrying...", name, serviceID)
			x.AddPublicNetworkToService(name, serviceID, nil)
		}
		return
	}
}

func (x *Docker) AddPublicNetwork(name string, containerID string, labels map[string]string) {
	if _, ok := labels["com.docker.swarm.service.name"]; ok {
		log.Println("[container] service container, skipping...", name, containerID)
		return
	}

	if _, ok := labels["proxy"]; !ok {
		log.Println("[container] container is not public, skipping...", name, containerID)
		return
	}

	serviceName := labels["com.docker.compose.service"]
	stackName := labels["com.docker.compose.project"]
	host := stackName + "_" + serviceName
	aliases := []string{host, serviceName} //FIXME: add serviceName only for service mode

	x.MapInternalPorts(host, serviceName, labels)

	// Skip network if not defined
	if x.networkID == "" {
		return
	}

	log.Println("[container] Adding container", name, "to network", x.networkName, aliases)
	options := network.EndpointSettings{}
	options.Aliases = aliases
	err := x.cli.NetworkConnect(x.ctx, x.networkID, containerID, &options)
	if err != nil {
		log.Errorln("[container] Fail to add container", name, "to network", x.networkName, ":", err)
		return
	}

}

func (x *Docker) MapInternalPorts(host, serviceName string, labels map[string]string) {
	if proxyHttpPort, ok := labels["proxy-http-port"]; ok {
		HttpPorts[host] = proxyHttpPort
		HttpPorts[serviceName] = proxyHttpPort //FIXME: add serviceName only for service mode
		log.Println("[mapping] Map http port for service", host, "/", serviceName, "to", proxyHttpPort)
	} else {
		delete(HttpPorts, host)
		delete(HttpPorts, serviceName) //FIXME: add serviceName only for service mode
	}
	if proxyTlsPort, ok := labels["proxy-tls-port"]; ok {
		TlsPorts[host] = proxyTlsPort
		TlsPorts[serviceName] = proxyTlsPort //FIXME: add serviceName only for service mode
		log.Println("[mapping] Map tls port for service", host, "/", serviceName, "to", proxyTlsPort)
	} else {
		delete(TlsPorts, host)
		delete(TlsPorts, serviceName) //FIXME: add serviceName only for service mode
	}
}

func (x *Docker) RemovePublicNetwork(name string, containerID string, labels map[string]string) {
	if _, ok := labels["com.docker.swarm.service.name"]; ok {
		return
	}

	if _, ok := labels["proxy"]; !ok {
		return
	}

	// Skip network if not defined
	if x.networkID == "" {
		return
	}

	log.Println("[container] Removing container", name, "from network", x.networkName)

	err := x.cli.NetworkDisconnect(x.ctx, x.networkID, containerID, true)
	if err != nil {
		log.Errorln("[container] Fail to remove container", name, " from network", x.networkName, ":", err)
	}
}

func (x *Docker) Events(msgchan <-chan events.Message, errchan <-chan error) {

	for {
		select {
		case msg := <-msgchan:
			log.Println("[event]", msg.Type, msg.Action, msg.Scope, msg.ID, msg.Status, msg.From)
			if msg.Type == events.ContainerEventType {
				log.Println("[event]", msg.Actor)
				if msg.Action == "start" {
					x.AddPublicNetwork(msg.Actor.Attributes["name"], msg.Actor.ID, msg.Actor.Attributes)
				} else if msg.Action == "die" {
					x.RemovePublicNetwork(msg.Actor.Attributes["name"], msg.Actor.ID, msg.Actor.Attributes)
				}
			} else if msg.Type == events.ServiceEventType {
				log.Println("[event]", msg.Actor)
				if msg.Action == "create" {
					x.AddPublicNetworkToService(msg.Actor.Attributes["name"], msg.Actor.ID, nil)
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

	log.Println("[docker] Listing Networks")
	networkListOptions := types.NetworkListOptions{}
	networks, err := x.cli.NetworkList(x.ctx, networkListOptions)
	if err != nil {
		panic(err)
	}

	x.networkID = ""
	if x.networkName != "" {
		for _, network := range networks {
			if network.Name == x.networkName {
				x.networkID = network.ID
			}
		}
		if x.networkID == "" {
			log.Fatalln("[docker] Public Network for docker not found :", x.networkName)
		}
	}
	log.Println("[docker] Public Networks", x.networkName, x.networkID)

	log.Println("[docker] Listening to docker events...")
	eventOptions := types.EventsOptions{}
	msgchan, errchan := x.cli.Events(x.ctx, eventOptions)

	log.Println("[docker] Listing containers...")
	containerListOptions := types.ContainerListOptions{}
	containers, _ := x.cli.ContainerList(x.ctx, containerListOptions)
	for _, container := range containers {
		log.Println(container.Names[0], container.ID, container.State, container.Labels)

		if container.State == "running" {
			x.RemovePublicNetwork(container.Names[0], container.ID, container.Labels)
			x.AddPublicNetwork(container.Names[0], container.ID, container.Labels)
		}
	}

	log.Println("[docker] Listing services...")
	serviceListoptions := types.ServiceListOptions{}
	services, _ := x.cli.ServiceList(x.ctx, serviceListoptions)
	for _, service := range services {
		log.Println(service.Spec.Name, service.ID, service.Spec.Labels)
		x.AddPublicNetworkToService(service.Spec.Name, service.ID, &service)
	}

	log.Println("[docker] Reading docker events...")
	go x.Events(msgchan, errchan)

	return &x
}
