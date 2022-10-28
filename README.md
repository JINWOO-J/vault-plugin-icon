# vault-plugin-icon


A HashiCorp Vault plugin that supports secp256k1 based signing, with an API interface that turns the vault into a software-based HSM device.

The plugin only exposes the following endpoints to enable the client to generate signing keys for the secp256k1 curve suitable for signing ICON transactions, <br> 
list existing signing keys by their names and addresses, and a `/sign` and `/param_sign` endpoint for each account. <br> 

It helps to generate and sign the private key in the Vault. <br> 
It never gives out the private keys. <br>

## Build

These dependencies are needed:

* go 1.18

To build the binary:
```bash
make 

####
darwin, amd64
[build-darwin-amd64]
[BUILD] build-darwin-amd64  / OS_ARCH=arm64, os=darwin arch=amd64, BIN=icon_darwin_amd64
github.com/JINWOO-J/vault-plugin-icon/backend
github.com/JINWOO-J/vault-plugin-icon
[BUILD] fdf50ce224998879bf6a1e972a242d9f4d848081a552476486ce4073c74d632e  ./build/icon_darwin_amd64

export SHASUM=fdf50ce224998879bf6a1e972a242d9f4d848081a552476486ce4073c74d632e
go  get
go  test  ./... -cover -coverprofile=coverage.txt -covermode=atomic
?   	github.com/JINWOO-J/vault-plugin-icon	[no test files]
ok  	github.com/JINWOO-J/vault-plugin-icon/backend	17.765s	coverage: 58.8% of statements

```

The output is `build/icon`

## Installing the Plugin on HashiCorp Vault server

The plugin must be registered and enabled on the vault server as a secret engine.

### Enabling on a dev mode server

Start the docker container 
```
cd docker
docker-compose -f docker-compose-local.dev.yml up -d

```

Build and enable plugin

```
make docker_dev
```


Run the example script

```
examples/create_sign_wallet.sh

```
