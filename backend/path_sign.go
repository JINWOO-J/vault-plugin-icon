// Copyright © 2022 Jinwoo
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package backend

import (
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

func pathSign(b *backend) *framework.Path {
	return &framework.Path{
		Pattern:      "accounts/" + framework.GenericNameRegex("name") + "/sign",
		HelpSynopsis: "Sign a provided transaction object.",
		HelpDescription: `

    Sign a transaction object with properties conforming to the ICON JSON-RPC documentation.

    `,
		Fields: map[string]*framework.FieldSchema{
			"name": &framework.FieldSchema{Type: framework.TypeString},
			"id": &framework.FieldSchema{
				Type:        framework.TypeInt,
				Description: "JSON RPC ID",
				Default:     2848,
			},
			"to": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "to address",
				Default:     "",
			},
			"data": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "The compiled code of a contract OR the hash of the invoked method signature and encoded parameters.",
			},
			"input": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "-",
			},
			"value": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "HEX of the value sent with this transaction ",
			},
			"nonce": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "The transaction nonce.",
			},
			"serialize": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "(optional) serialize of the target blockchain network.",
				Default:     "",
			},
			"params": &framework.FieldSchema{
				Type:        framework.TypeMap,
				Description: "(optional) params of the target blockchain network. ",
				Default:     "",
			},
		},
		ExistenceCheck: b.pathExistenceCheck,
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.CreateOperation: &framework.PathOperation{
				Callback: b.signTx,
			},
		},
	}
}

func pathParamSign(b *backend) *framework.Path {
	return &framework.Path{
		Pattern:      "accounts/" + framework.GenericNameRegex("from") + "/param_sign",
		HelpSynopsis: "Sign a provided transaction object.",
		HelpDescription: `

    Sign a transaction object with properties conforming to the ICON JSON-RPC documentation.

    `,
		Fields: map[string]*framework.FieldSchema{
			"from": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "From address, It is forcibly converted to the registered account name.",
			},
			"to": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "to address",
				//Default:     "",
			},
			"data": &framework.FieldSchema{
				Type:        framework.TypeMap,
				Description: "The compiled code of a contract OR the hash of the invoked method signature and encoded parameters.",
			},
			"value": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "HEX of the value sent with this transaction ",
			},
			"nonce": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "The transaction nonce.",
			},
			"stepLimit": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "The stepLimit(Fee) price for the transaction.",
				//Default:     "",
			},
			"nid": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "(optional) Network ID of the target blockchain network",
				Default:     "0x1",
			},
			"serialize": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "(optional) serialize of the target blockchain network.",
				Default:     "",
			},
			"timestamp": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Timestamp",
				Default:     TimeStampNow(),
			},
		},
		ExistenceCheck: b.pathExistenceCheck,
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.CreateOperation: &framework.PathOperation{
				Callback: b.signTransaction,
			},
		},
	}
}

func pathSignAuth(b *backend) *framework.Path {
	return &framework.Path{
		Pattern:      "accounts/" + framework.GenericNameRegex("walletAddress") + "/sign_auth",
		HelpSynopsis: "Sign a provided transaction object.",
		HelpDescription: `

    Sign a transaction object with getting Planet manager's credentials.

    `,
		Fields: map[string]*framework.FieldSchema{
			"walletAddress": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "walletAddress, It is forcibly converted to the registered account name.",
			},
			"time": &framework.FieldSchema{
				Type:        framework.TypeInt,
				Description: "Timestamp seconds (INT)",
				Default:     0,
			},
		},
		ExistenceCheck: b.pathExistenceCheck,
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.CreateOperation: &framework.PathOperation{
				Callback: b.signAuth,
			},
		},
	}
}
