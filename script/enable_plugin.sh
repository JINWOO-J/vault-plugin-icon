#!/bin/bash
export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_BIN=${VAULT_BIN:-"./vault"}
export DIR="$(cd "$(dirname "$(readlink "$0")")" && pwd)"
export VAULT_DIR=${VAULT_DIR:-"${DIR}"}
export PLUGIN_NAME="iconsign"

if [[ $DIR == *"script" ]]; then
  echo "It's there!"
    BUILD_PATH=".."
else
    BUILD_PATH="."
fi

echo "      BUILD Vault plugin"
go  build -o ${PLUGIN_NAME} \
    -ldflags "-X main.buildDate=`date -u +\"%Y-%m-%dT%H:%M:%SZ\"` -X main.buildVersion=" \
    -tags=prod -v $BUILD_PATH || exit

cd ${VAULT_DIR}

SHASUM=$(shasum -a 256 "plugin/${PLUGIN_NAME}" | cut -d " " -f1)
echo "    Registering plugin - ${SHASUM}"
$VAULT_BIN write sys/plugins/catalog/${PLUGIN_NAME} \
  sha_256="$SHASUM" \
  command="iconsign"


echo "    Mounting plugin"
$VAULT_BIN secrets enable -path=icon -description="ICON Wallet" -plugin-name=iconsign plugin




#$VAULT_BIN write icon/accounts privateKey=37677f06d3d58c1c62c9dd6a8e7e3e57178e90591ac5f8d844aaf991db3face2


#./vault write icon/accounts/hxb87f5bbcc4e5490d636928f93706baa8ef6bd7d0/param_sign data='{
#  "id": 2848,
#  "jsonrpc": "2.0",
#  "method": "icx_sendTransaction",
#  "params": {
#    "from": "hx5a05b58a25a1e5ea0f1d5715e1f655dffc1fb30a",
#    "to": "hx32b5704b766c535c34291c0d10ddd5bbd7b6b9fb",
#    "stepLimit": "0x4a817c800",
#    "value": "0x38d7ea4c68000",
#    "nid": "0x53",
#    "nonce": "0x8",
#    "version": "0x3",
#    "timestamp": "0x5e5dfa3af07d0"
#  }
#}'

#curl -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" http://localhost:8200/v1/icon/accounts/hxb87f5bbcc4e5490d636928f93706baa8ef6bd7d0/sign -d '{
#  "id": 2848,
#  "jsonrpc": "2.0",
#  "method": "icx_sendTransaction",
#  "params": {
#    "from": "hx5a05b58a25a1e5ea0f1d5715e1f655dffc1fb30a",
#    "to": "hx32b5704b766c535c34291c0d10ddd5bbd7b6b9fb",
#    "stepLimit": "0x4a817c800",
#    "value": "0x38d7ea4c68000",
#    "nid": "0x53",
#    "nonce": "0x8",
#    "version": "0x3",
#    "timestamp": "0x5e5dfa3af07d0"
#  }
#}'| jq





