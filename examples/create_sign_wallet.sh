#!/bin/bash

DEV_TOKEN="root"

echo "==== Create Wallet ==="
curl -s -H "Content-Type: application/json" -H "Authorization: Bearer $DEV_TOKEN" \
    -X POST -d '{"name": "test_wallet"}' \
    http://localhost:8200/v1/icon/accounts

curl -s -H "Content-Type: application/json" -H "Authorization: Bearer $DEV_TOKEN" \
    -X POST -d '{"name": "test_wallet_private_key", "privateKey": "506e1e6c40fb1d2659b893c62409b667f9880df85d748874ebc3d9e03d33e7a2"}' \
    http://localhost:8200/v1/icon/accounts


echo ""
echo "=== List Wallet ==="
curl -s -H "Content-Type: application/json" -H "Authorization: Bearer $DEV_TOKEN" \
    http://localhost:8200/v1/icon/accounts?list=true |jq


echo ""
echo "1. === Sign payload ==="
curl -s -H "Content-Type: application/json" -H "Authorization: Bearer $DEV_TOKEN" \
    http://localhost:8200/v1/icon/accounts/hx583769035c6f7b86231b6f1f5a2f545e4a208204/sign \
    -X POST -d '{
                 "id": 2848,
                 "jsonrpc": "2.0",
                 "method": "icx_sendTransaction",
                 "params": {
                   "from": "hx5a05b58a25a1e5ea0f1d5715e1f655dffc1fb30a",
                   "to": "hx32b5704b766c535c34291c0d10ddd5bbd7b6b9fb",
                   "stepLimit": "0x4a817c800",
                   "value": "0x38d7ea4c68000",
                   "nid": "0x53",
                   "nonce": "0x8",
                   "version": "0x3",
                   "timestamp": "0x5e5dfa3af07d0"
                 }
               }' | jq

echo ""
echo "2. === Sign payload ==="
curl -s -H "Content-Type: application/json" -H "Authorization: Bearer $DEV_TOKEN" \
    http://localhost:8200/v1/icon/accounts/hx583769035c6f7b86231b6f1f5a2f545e4a208204/param_sign \
    -X POST -d '{
                   "from": "hx5a05b58a25a1e5ea0f1d5715e1f655dffc1fb30a",
                   "to": "hx32b5704b766c535c34291c0d10ddd5bbd7b6b9fb",
                   "stepLimit": "0x4a817c800",
                   "value": "0x38d7ea4c68000",
                   "nid": "0x53",
                   "nonce": "0x8",
                   "version": "0x3",
                   "timestamp": "0x5e5dfa3af07d0"
               }' | jq
