#!/bin/bash
export VAULT_ADDR='http://127.0.0.1:8200'

PLUGIN_DIR="vault-plugin-icon"
PLUGIN_NAME="icon"


cd $PLUGIN_DIR

#make all|| exit
go  build -o ${PLUGIN_NAME} -ldflags "-X main.buildDate=`date -u +\"%Y-%m-%dT%H:%M:%SZ\"` -X main.buildVersion=" -tags=prod -v || exit

cd -

cp -rf ${PLUGIN_DIR}/${PLUGIN_NAME} ./plugin/

SHASUM=$(shasum -a 256 "${PLUGIN_DIR}/${PLUGIN_NAME}" | cut -d " " -f1)
echo $SHASUM

echo "    Registering plugin - ${SHASUM}"


#./vault write sys/plugins/catalog/iconsign \
#  sha_256="$SHASUM" \
#  command="iconsign"



./vault plugin register -sha256=${SHASUM} iconsign


#echo "    Reloading plugin"
#./vault write sys/plugins/reload/backend \
#  plugin="iconsign"
