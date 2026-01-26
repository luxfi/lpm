// Copyright (C) 2019-2025, Lux Partners Limited. All rights reserved.
// See the file LICENSE for licensing terms.

package admin

import (
	"context"

	"github.com/luxfi/sdk/admin"
)

var _ Client = &client{}

type Client interface {
	LoadVMs() error
	WhitelistChain(chainID string) error
}

type client struct {
	client *admin.Client
}

func NewClient(url string) Client {
	return &client{
		client: admin.NewClient(url),
	}
}

func (c *client) LoadVMs() error {
	_, _, err := c.client.LoadVMs(context.Background())

	return err
}

func (c *client) WhitelistChain(chainID string) error {
	// id, err := ids.FromString(chainID)
	// if err != nil {
	// 	return err
	// }
	//
	// _, err = c.client.WhitelistChain(context.Background(), id)
	// return err
	return nil
}
