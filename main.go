package main

import (
	"log"

	"github.com/fsouza/go-dockerclient"
)

func main() {
	// setup the docker event listening
	endpoint := "unix:///var/run/docker.sock"
	client, _ := docker.NewClient(endpoint)
	v := IntVault{}
	v.NewVaultClient()

	// init channel to retreive docker events
	evChan := make(chan *docker.APIEvents)
	// assign our channel to be the event listener
	client.AddEventListener(evChan)

	// range over the channel outputing only if the action is start
	// ever time a new container starts we need to do the vault token dance
	for elem := range evChan {
		if elem.Action == "start" {
			// fmt.Printf("%+v", elem)
			log.Println(elem.Actor.Attributes["name"], "Has Started")
			v.createPermToken(elem.Actor.Attributes["name"])
			v.createTempToken(elem.Actor.Attributes["name"])
			tokens := make(map[string]string)
			tokens["temp"] = v.tempToken
			tokens["perm"] = v.permToken
			tokens["accessor"] = v.permAccessor

			// save temp token
			persistData(fileLocation(), elem.Actor.Attributes["name"], tokens)
		}
	}

}
