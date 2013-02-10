// Copyright 2013, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// NOTE FOR IMPLEMENTING API CALLS
// Each API call gets a type and has a hidden *Client attribute.
// The API call constructors are methods of *Client.
// Each API call satisfies the Request interface
// Each API call defines an Exec method that calls call.client.Do(call)
// Each API call the response of each API call can be different

package s3

/*  Filename:    s3.go
 *  Author:      Bryan Matsuo <bryan.matsuo [at] gmail.com>
 *  Created:     2013-01-23 23:47:58.358434 -0800 PST
 *  Description: 
 */

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/bmatsuo/go-aws"
)

var (
	USStandard   = &aws.Region{"https", "s3.amazonaws.com"}
	USWest1      = &aws.Region{"https", "s3-us-west-1.amazonaws.com"}
	USWest2      = &aws.Region{"https", "s3-us-west-2.amazonaws.com"}
	EUWest1      = &aws.Region{"https", "s3-eu-west-1.amazonaws.com"}
	APSouthEast1 = &aws.Region{"https", "s3-as-southeast-1.amazonaws.com"}
	APSouthEast2 = &aws.Region{"https", "s3-as-southeast-2.amazonaws.com"}
	APNorthEast2 = &aws.Region{"https", "s3-as-northeast-2.amazonaws.com"}
	SAEast1      = &aws.Region{"https", "s3-sa-east-1.amazonaws.com"}
)

type Client struct {
	*aws.Credentials
	*aws.Region
	client *http.Client
}

func NewClient(creds *aws.Credentials, region *aws.Region) *Client {
	return &Client{
		Credentials: creds,
		Region:      region,
		client:      &http.Client{},
	}
}

type Request interface {
	Request(*aws.Region) (*http.Request, error)
}

type Error struct {
	Code              string
	Message           string
	RequestId, HostId string
	Header            string
	StringToSignBytes string
	SignatureProvided string
}

func (err *Error) StringToSign() string {
	var p []byte
	for _, cstr := range strings.Split(err.StringToSignBytes, " ") {
		c, _ := strconv.ParseInt(cstr, 16, 8)
		p = append(p, byte(c))
	}
	return string(p)
}

func (err *Error) Error() string {
	return fmt.Sprintf("%#v", err)
}

// Creates a signed request from req
func (client *Client) Request(req Request) (*http.Request, error) {
	hreq, err := req.Request(client.Region)
	if err != nil {
		return nil, err
	}
	client.Sign(hreq)
	return hreq, nil
}

func (client *Client) Do(req Request) (*http.Response, error) {
	hreq, err := client.Request(req)
	if err != nil {
		return nil, err
	}
	resp, err := client.client.Do(hreq)
	if err != nil {
		return nil, err
	}
	if resp.Header.Get("Content-Type") == "application/xml" && resp.StatusCode >= 300 && resp.StatusCode < 600 {
		var err Error
		body, _ := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		xml.Unmarshal(body, &err)
		return resp, &err
	}
	return resp, nil
}

// REST authentication
func (client *Client) Sign(req *http.Request) {
	now := time.Now().UTC().Format(time.RFC1123)
	req.Header.Set("Date", now)
	req.Header.Set("x-amz-date", now)
	tosign := req.Method
	tosign += "\n"
	tosign += req.Header.Get("Content-MD5")
	tosign += "\n"
	tosign += req.Header.Get("Content-Type")
	tosign += "\n"
	//tosign += now
	tosign += "\n"
	headers := make(normalizedHeaders, 0, len(req.Header))
	for name, value := range req.Header {
		headers = append(headers, newNormalizedHeader(name, value))
	}
	sort.Sort(headers)
	for _, h := range headers {
		if !h.IsAWS() {
			continue
		}
		tosign += h.Normal
		tosign += ":"
		tosign += h.Value
		tosign += "\n"
	}

	resource := req.URL.Path
	if req.URL.RawQuery != "" {
		// ?acl and ?torrent (http://bit.ly/pO46SK)
		resource += "?"
		resource += req.URL.RawQuery
	}
	tosign += resource

	auth := "AWS"
	auth += " "
	auth += client.Credentials.AccessKeyId
	auth += ":"
	auth += client.signature(tosign)
	req.Header.Set("Authorization", auth)
}

// Query string authentication
func (client *Client) SignUrl(req *url.URL, lifetime int64) {
	tosign := "GET"
	tosign += "\n"
	tosign += "\n" // no content-type/md5 for GET requests
	tosign += "\n"
	expires := strconv.FormatInt(time.Now().Unix()+lifetime, 10)
	tosign += expires
	tosign += "\n"
	tosign += req.Path

	query := "AWSAccessKeyId="
	query += client.Credentials.AccessKeyId
	query += "&Expires="
	query += expires
	query += "&Signature="
	query += string(client.signature(tosign))

	req.RawQuery = query
}
func (client *Client) signature(tosign string) string {
	hash := hmac.New(sha1.New, []byte(client.Credentials.SecretAccessKey))
	hash.Write([]byte(tosign))
	sum := hash.Sum(make([]byte, 0, 50))
	signature := make([]byte, base64.URLEncoding.EncodedLen(len(sum)))
	base64.StdEncoding.Encode(signature, sum)
	return string(signature)
}

type baseRequest struct {
	Method string
	Path   string
	Query  url.Values
	Header http.Header
	Body   io.ReadCloser
}

type normalizedHeaders []*normalizedHeader
type normalizedHeader struct {
	Normal string
	Header string // original header
	Value  string // trimmed of concatenated
}

func newNormalizedHeader(header string, value []string) *normalizedHeader {
	h := &normalizedHeader{
		Normal: strings.ToLower(string(header)),
		Header: header,
	}

	values := make([]string, 0, len(value))
	for i := range value {
		normal := strings.TrimFunc(value[i], unicode.IsSpace)
		values = append(values, normal)
	}
	h.Value = strings.Join(values, ",")

	return h
}

func (h normalizedHeader) IsAWS() bool {
	return strings.HasPrefix(h.Normal, "x-amz")
}

func (hs normalizedHeaders) Len() int           { return len(hs) }
func (hs normalizedHeaders) Less(i, j int) bool { return hs[i].Normal < hs[j].Normal } // TODO maybe not completely right
func (hs normalizedHeaders) Swap(i, j int)      { hs[i], hs[j] = hs[j], hs[i] }
