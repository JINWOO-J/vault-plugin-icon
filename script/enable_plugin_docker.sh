#!/bin/sh
export VAULT_BIN=${VAULT_BIN:-"vault"}
export VAULT_DIR=${VAULT_DIR:-"/vault"}
export PLUGIN_DIR=${PLUGIN_DIR:-"/vault/plugin"}
export OS=$(uname | tr A-Z a-z)
export ARCH=$(arch)

if [ ${ARCH} == "x86_64" ]; then
    ARCH="amd64"
fi

export PLUGIN_NAME=${PLUGIN_NAME:-"icon"}
export PLUGIN_BIN_FILE=${PLUGIN_BIN_FILE:-"${PLUGIN_NAME}_${OS}_${ARCH}"}

echo $PLUGIN_BIN_FILE

export SHA3SUM=$(sha256sum "${PLUGIN_DIR}/${PLUGIN_BIN_FILE}" | cut -d " " -f1)

echo "    Registering plugin - ${SHA3SUM}"
$VAULT_BIN write sys/plugins/catalog/${PLUGIN_NAME} \
  sha_256="$SHA3SUM" \
  command="${PLUGIN_BIN_FILE}"

echo "    Mounting plugin"
$VAULT_BIN secrets enable -path=icon -description="ICON Wallet Signer" -plugin-name=${PLUGIN_NAME} plugin

echo "    Reload plugin"
$VAULT_BIN write sys/plugins/reload/backend plugin=${PLUGIN_NAME}


#$VAULT_BIN write icon/accounts privateKey=506e1e6c40fb1d2659b893c62409b667f9880df85d748874ebc3d9e03d33e7a2


#./vault write icon/accounts/hx583769035c6f7b86231b6f1f5a2f545e4a208204/param_sign data='{
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
#    "from": "hx583769035c6f7b86231b6f1f5a2f545e4a208204",
#    "to": "hx32b5704b766c535c34291c0d10ddd5bbd7b6b9fb",
#    "stepLimit": "0x4a817c800",
#    "value": "0x38d7ea4c68000",
#    "nid": "0x53",
#    "nonce": "0x8",
#    "version": "0x3",
#    "timestamp": "0x5e5dfa3af07d0"
#  }
#}'| jq





