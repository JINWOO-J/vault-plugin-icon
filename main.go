package main

import (
	"github.com/JINWOO-J/vault-plugin-icon/backend"
	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/sdk/plugin"
	"log"
	"os"
)

var (
	buildDate    = ""
	buildVersion = ""
)

func main() {
	log.Println("buildDate:", buildDate, ", buildVersion:", buildVersion)
	apiClientMeta := &api.PluginAPIClientMeta{}
	flags := apiClientMeta.FlagSet()
	flags.Parse(os.Args[1:]) // Ignore command, strictly parse flags

	tlsConfig := apiClientMeta.GetTLSConfig()
	tlsProviderFunc := api.VaultPluginTLSProvider(tlsConfig)

	err := plugin.Serve(&plugin.ServeOpts{
		BackendFactoryFunc: backend.Factory,
		TLSProviderFunc:    tlsProviderFunc,
	})
	if err != nil {
		log.Println("=== ERROR ===", err)
		os.Exit(1)
	}
}
