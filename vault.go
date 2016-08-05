package main

import (
	"fmt"
	"log"

	vaultapi "github.com/hashicorp/vault/api"
)

type IntVault struct {
	client            *vaultapi.Client
	wrapInfo          *vaultapi.SecretWrapInfo
	tempToken         string
	permToken         string
	permAccessor      string
	selfTokenToRevoke string
	selfActiveToken   string
}

// TODO: @debug this is likely to be deprecated
// func defaultWrappingLookupFunc(operation, path string) string {
// 	// return os.Getenv(vaultapi.EnvVaultWrapTTL)
// 	return "30s"
// }

func (v *IntVault) NewVaultClient() {
	var err error
	v.client, err = vaultapi.NewClient(vaultapi.DefaultConfig())
	if err != nil {
		log.Println(err)
	}
	// set the wrap ttl
	// v.client.SetWrappingLookupFunc(defaultWrappingLookupFunc)
}

func (v *IntVault) tokenLookup(token string) {
	// Check for an empty token. If empty just lookup self
	if token == "self" {
		// lookup self and safe results as selfTokenToRevoke
		selfSecret, err := v.client.Auth().Token().LookupSelf()
		if err != nil {
			log.Println(err)
		}
		// fmt.Printf("%+v\n", selfSecret)
		v.setSelfAccessorToRevoke(selfSecret)
	} else {
		v.client.Auth().Token().Lookup(token)
	}

}

func (v *IntVault) setSelfAccessorToRevoke(secret *vaultapi.Secret) {
	v.selfTokenToRevoke = secret.Data["accessor"].(string)
}

func (v *IntVault) createSelfToken() {

	policies := []string{"root"}

	// generate token
	secret, err := v.client.Auth().Token().Create(&vaultapi.TokenCreateRequest{
		NumUses:     0,
		DisplayName: "TULLY",
		Policies:    policies,
		NoParent:    true,
	})
	if err != nil {
		log.Println(err)
	}

	// fmt.Printf("%+v\n", secret)
	fmt.Println("My New Token", secret.Auth.ClientToken)

	v.selfActiveToken = secret.Auth.ClientToken

}

func (v *IntVault) createPermToken(requestor string) {

	policies := []string{"testapi_user"}

	// generate token and stick cubby hole
	secret, err := v.client.Auth().Token().Create(&vaultapi.TokenCreateRequest{
		NumUses:     0,
		DisplayName: requestor,
		Policies:    policies,
	})
	if err != nil {
		log.Println(err)
	}

	// fmt.Printf("%+v\n", secret)
	fmt.Println("Perm Token", secret.Auth.ClientToken)
	fmt.Println("Perm Accessor", secret.Auth.Accessor)
	// fmt.Printf("%+v\n", secret.WrapInfo)
	v.permToken = secret.Auth.ClientToken
	v.permAccessor = secret.Auth.Accessor

}

func (v *IntVault) createTempToken(requestor string) {

	policies := []string{"testapi_user"}

	// generate temp to auth cubby hole
	secret, err := v.client.Auth().Token().Create(&vaultapi.TokenCreateRequest{
		NumUses:     3,
		DisplayName: requestor,
		Policies:    policies,
		TTL:         "60s",
	})
	if err != nil {
		log.Println(err)
	}

	v.tempToken = secret.Auth.ClientToken
	fmt.Println("Temp Token", v.tempToken)

	writeToCubby(v.tempToken, v.permToken, requestor)

}

func writeToCubby(temp, perm, appName string) {
	client, err := vaultapi.NewClient(vaultapi.DefaultConfig())
	if err != nil {
		log.Println(err)
	}
	client.SetToken(temp)
	l := client.Logical()
	_, err = l.Write("cubbyhole/"+appName,
		map[string]interface{}{
			"value": perm,
		})
	if err != nil {
		panic(err)
	}

}

func (v *IntVault) revokeAccessor(accessor string) {
	log.Println("Going to revoke with accessor", accessor)
	v.client.Auth().Token().RevokeAccessor(accessor)
}
