package tea

import (
	vaultapi "github.com/hashicorp/vault/api"
)

type Vault struct {
	client *vaultapi.Client
}

func NewVault() (*Vault, error) {
	client, err := vaultapi.NewClient(vaultapi.DefaultConfig())
	if err != nil {
		return nil, err
	}

	return &Vault{client: client}, nil
}
