package entity

import (
	"fmt"

	"github.com/photoprism/photoprism/internal/form"
)

// AddClient creates a new client and returns it if successful.
func AddClient(frm form.Client) (client *Client, err error) {
	if found := FindClientByUID(frm.ID()); found != nil {
		return found, fmt.Errorf("client id %s already exists", found.ClientUID)
	}

	client = NewClient().SetFormValues(frm)

	if err = client.Validate(); err != nil {
		return client, err
	} else if err = client.Create(); err != nil {
		return client, err
	}

	return client, nil
}
