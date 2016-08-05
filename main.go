package main

import (
	"log"
	"os"

	"github.com/fsouza/go-dockerclient"
)

// TODO: create mechanism for renewing and revoking my own tokens
// generate tokens in the same way and write them out and revoke etc.
// always be sure to read in new token before performing actions
// my temp tokens need to not have a ttl so they can sit waiting for a time
// that a container is launched on said host and can then use that token

func main() {
	log.Println("TULLY Keymaster service starting")
	// setup the docker event listening
	endpoint := "unix:///var/run/docker.sock"
	client, _ := docker.NewClient(endpoint)
	v := IntVault{}
	v.NewVaultClient()
	v.selfActiveToken = os.Getenv("VAULT_TOKEN")

	// init channel to retreive docker events
	evChan := make(chan *docker.APIEvents)
	// assign our channel to be the event listener
	client.AddEventListener(evChan)

	// range over the channel outputing only if the action is start
	// ever time a new container starts we need to do the vault token dance
	for elem := range evChan {
		if elem.Action == "start" {
			// v.NewVaultClient()
			v.client.SetToken(v.selfActiveToken)
			log.Println("Using", v.selfActiveToken, "As the TULLY Token for this loop")
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
			// find my current accessor
			v.tokenLookup("self")
			v.tokenLookup(v.selfActiveToken)

			v.createSelfToken()
			log.Println("My new Token is", v.selfActiveToken)
			// set token. This could be what we aread form ENV vars
			// or what it has been reset to since first run

			// TODO: this new token isn't actually being used need to figure this out
			// when I start a second container its still using the token from the first run not the second one
			// v.client.SetToken(v.selfActiveToken)
			v.revokeAccessor(v.selfTokenToRevoke)

		}
		if elem.Action == "stop" {
			// TODO: if a stop event is observed cleanup files and revoke tokens for
			// the stopped container
		}
	}

}
