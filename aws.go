// Copyright 2013, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*  Filename:    aws.go
 *  Author:      Bryan Matsuo <bryan.matsuo [at] gmail.com>
 *  Created:     2013-01-23 23:36:58.888125 -0800 PST
 *  Description: Main source file in go-aws
 */

// Package aws does ....
package aws

import (
	"net/url"
	"os"
)

type Credentials struct {
	AccessKeyId     string
	SecretAccessKey string
}

func Getenv() *Credentials {
	creds := &Credentials{
		AccessKeyId:     os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
	}
	return creds
}

// Each aws product should define it's own region vars in a consistent manner.
type Region struct {
	Protocol string
	Endpoint string
}

func (region *Region) Url(subdomain, path string, query url.Values) *url.URL {
	host := region.Endpoint
	if subdomain != "" {
		host = subdomain + "." + host
	}
	qstr := ""
	if query != nil {
		qstr = query.Encode()
	}
	return &url.URL{
		Scheme:   region.Protocol,
		Host:     host,
		Path:     path,
		RawQuery: qstr,
	}
}
