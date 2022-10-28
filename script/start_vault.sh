#!/bin/bash
#set -e
export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_BIN=${VAULT_BIN:-"./vault"}
export DIR="$(cd "$(dirname "$(readlink "$0")")" && pwd)"
export VAULT_DIR=${VAULT_DIR:-"${DIR}"}
export PLUGIN_NAME=${PLUGIN_NAME:-"iconsign"}


if [[ $DIR == *"script" ]]; then
  echo "It's there!"
    BUILD_PATH=".."
else
    BUILD_PATH="."
fi

echo "VAULT_BIN=${VAULT_BIN} | VAULT_DIR=${VAULT_DIR} | BUILD_PATH=${BUILD_PATH}"

echo "      BUILD Vault plugin"
go  build -o ${PLUGIN_NAME} \
    -ldflags "-X main.buildDate=`date -u +\"%Y-%m-%dT%H:%M:%SZ\"` -X main.buildVersion=" \
    -tags=prod -v $BUILD_PATH || exit

file ${BUILD_PATH}/${PLUGIN_NAME} || exit;
echo "built=${BUILD_PATH}/${PLUGIN_NAME}, bin=${VAULT_DIR}/${PLUGIN_NAME} plugin=${VAULT_DIR}/plugin"


mkdir -p ${VAULT_DIR}/plugin
cp -rf ${BUILD_PATH}/${PLUGIN_NAME} ${VAULT_DIR}/plugin/
cp -f ./script/config.hcl ${VAULT_DIR}/

cd ${VAULT_DIR}

pwd

echo "    Starting"
$VAULT_BIN server \
  -dev \
  -dev-root-token-id="root" \
  -log-level="debug" \
  -dev-plugin-dir=./plugin \
  -dev-listen-address="0.0.0.0:8200" \
&
sleep 2
VAULT_PID=$!

#  -config="config.hcl" \
#  -config="$SCRATCH/vault.hcl" \

echo "----------------"

function cleanup {
  echo ""
  echo "==> Cleaning up"
  kill -INT "$VAULT_PID"
  rm -rf "$SCRATCH"
}
trap cleanup EXIT


SHASUM=$(shasum -a 256 "plugin/${PLUGIN_NAME}" | cut -d " " -f1)

echo "    Registering plugin - ${SHASUM}"

$VAULT_BIN write sys/plugins/catalog/iconsign sha_256="$SHASUM" command="iconsign"


#
echo "    Mounting plugin"
$VAULT_BIN secrets enable -path=icon -description="ICON Wallet" -plugin-name=iconsign plugin


VAULT_VERSION=$(${VAULT_BIN} --version)

echo "==> Ready to Vault! Ver = ${VAULT_VERSION}"
wait $!
