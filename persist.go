package main

import (
	"io/ioutil"
	"os"
)

func fileLocation() string {
	envLocation := os.Getenv("TULLY_PERSIST")
	if envLocation == "" {
		return "/tmp/"
	} else {
		return os.Getenv("TULLY_PERSIST")
	}

}

func makeAppDir(path, appName string) string {
	persistPath := path + appName

	// look to see if the dir exists. If it does just return the path
	if _, err := os.Stat(persistPath); err == nil {
		return persistPath
	}

	// if the dir does not exist continue on creating it
	err := os.Mkdir(path+appName, 0700)
	if err != nil {
		panic(err)
	}

	return path + appName

}

func readLocalAccessor(path string) string {
	dat, err := ioutil.ReadFile(path + "/accessor")
	if err != nil {
		panic(err)
	}
	return (string(dat))
}

func persistData(path, appName string, tokens map[string]string) {

	// write the dir for this application and change the path var to be our new full path with app name
	path = makeAppDir(path, appName)

	// write the app token
	err := ioutil.WriteFile(path+"/token", []byte(tokens["perm"]), 0600)
	if err != nil {
		panic(err)
	}

	// revoke previous token via its accessor
	if _, err := os.Stat(path + "/accessor"); err == nil {
		v := IntVault{}
		v.NewVaultClient()
		v.revokeAccessor(readLocalAccessor(path))
	}

	// write the accesor token for the app token
	err = ioutil.WriteFile(path+"/accessor", []byte(tokens["accessor"]), 0600)
	if err != nil {
		panic(err)
	}

}
