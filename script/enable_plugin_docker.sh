#!/bin/sh
export VAULT_BIN=${VAULT_BIN:-"vault"}
export VAULT_DIR=${VAULT_DIR:-"/vault"}
export PLUGIN_DIR=${PLUGIN_DIR:-"/vault/plugin"}
export OS=$(uname | tr A-Z a-z)
export ARCH=$(arch)

if [ ${ARCH} == "x86_64" ]; then
    ARCH="amd64"
elif [ ${ARCH} == "aarch64" ]; then
    ARCH="arm64"
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
