// Copyright (C) 2019-2021, Lux Partners Limited. All rights reserved.
// See the file LICENSE for licensing terms.

package admin

import (
	"context"

	adminapi "github.com/luxfi/node/api/admin"
)

var _ Client = &client{}

type Client interface {
	LoadVMs() error
	WhitelistSubnet(subnetID string) error
}

type client struct {
	client *adminapi.Client
}

func NewClient(url string) Client {
	return &client{
		client: adminapi.NewClient(url),
	}
}

func (c *client) LoadVMs() error {
	_, _, err := c.client.LoadVMs(context.Background())

	return err
}

func (c *client) WhitelistSubnet(subnetID string) error {
	// id, err := ids.FromString(subnetID)
	// if err != nil {
	// 	return err
	// }
	//
	// _, err = c.client.WhitelistSubnet(context.Background(), id)
	// return err
	return nil
}
