/*
	Copyright (C) CESS. All rights reserved.
	Copyright (C) Cumulus Encrypted Storage System. All rights reserved.

	SPDX-License-Identifier: Apache-2.0
*/

package client

func (c *Cli) Exit(role string) (string, error) {
	return c.Chain.Exit(role)
}