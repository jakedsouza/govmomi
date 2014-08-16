/*
Copyright (c) 2014 VMware, Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cli

import (
	"errors"
	"flag"
	"net/url"
	"os"

	"github.com/vmware/govmomi"
)

const cDescr = "ESX or vCenter URL"

type Client struct {
	*govmomi.Client
	u *url.URL
}

func (c *Client) String() string {
	if c.u != nil {
		withoutCredentials := *c.u
		withoutCredentials.User = nil
		return withoutCredentials.String()
	}
	return ""
}

func (c *Client) Set(s string) error {
	var err error

	c.u, err = url.Parse(s)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Register(f *flag.FlagSet) {
	c.Set(os.Getenv("GOVMOMI_URL"))
	f.Var(c, "u", cDescr)
}

func (c *Client) Process() error {
	var err error

	if c.u == nil {
		return errors.New("specify an " + cDescr)
	}

	c.Client, err = govmomi.NewClient(*c.u)
	if err != nil {
		return err
	}

	return nil
}
