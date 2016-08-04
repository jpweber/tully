package main

import (
	"log"

	"github.com/fsouza/go-dockerclient"
)

// TODO: create mechanism for renewing and revoking my own tokens
// generate tokens in the same way and write them out and revoke etc.
// always be sure to read in new token before performing actions
// my temp tokens need to not have a ttl so they can sit waiting for a time
// that a container is launched on said host and can then use that token

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

			// rotate my own tokens
			v.tokenLookup("self")
			log.Println("My current active token that I will revoke is", v.selfTokenToRevoke)
		}
	}

}
