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

func pathCreateAndList(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "accounts/?",
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ListOperation: &framework.PathOperation{
				Callback: b.listAccounts,
			},
			logical.UpdateOperation: &framework.PathOperation{
				Callback: b.createAccount,
			},
		},

		HelpSynopsis: "List all the ICON accounts maintained by the plugin backend and create new accounts.",
		HelpDescription: `

    LIST - list all accounts
    POST - create a new account

    `,
		Fields: map[string]*framework.FieldSchema{
			"privateKey": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Hexidecimal string for the private key (32-byte or 64-char long). If present, the request will import the given key instead of generating a new key.",
				Default:     "",
			},
			"name": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Alias ​​address of the wallet",
				Default:     "",
			},
			"detail": &framework.FieldSchema{
				Type:        framework.TypeBool,
				Description: "Choose the detail options (true/false)",
				Default:     true,
			},
		},
	}
}
