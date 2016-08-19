package main

import (
	"log"
	"os"

	"github.com/fsouza/go-dockerclient"
)

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
			v.revokeAccessor(v.selfTokenToRevoke)

		}
		if elem.Action == "stop" {
			// TODO: if a stop event is observed cleanup files and revoke tokens for
			// the stopped container
			// log.Println(elem.Actor.Attributes["name"], "Has Stopped")
			// // read in accessor token of application
			// filePath := fileLocation()
			// filePath = filePath + elem.Actor.Attributes["name"]
			// appAccessor := readLocalAccessor(fileLocation())

			// // revoke token using accessor
			// revokeSuccess := v.revokeAccessor(appAccessor)

			// // on success of token revocation delete the directory container the token and accessor files
			// if revokeSuccess {
			// 	err := os.Remove(filePath)
			// 	if err != nil {
			// 		log.Println("Error removing app token dir", err)
			// 	}
			// }
		}
	}

}
