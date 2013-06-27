// Copyright 2013, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// ses.go [created: Wed, 26 Jun 2013]

/*
a wrapper around the aws ses api.

	creds := &aws.Credentials{AccessKeyId, SecretAccessKey}
	result, err := ses.NewSendEmailRequest().
		To("exàmple@example.com").
		Source("noreply@example.com").
		Subject("Welcome").
		Text("Hello, example!").
		Html("<h1>Hello, example!</h1>").
		Exec(creds)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Sent email (SES MessageId %s)", result.MessageId)

*/
package ses

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/bmatsuo/go-aws"
)

type SendEmailResult struct {
	MessageId string `xml:"SendEmailResult>MessageId"`
	*ResponseMetadata
}

type ResponseMetadata struct {
	RequestId string
}

type SendEmailRequest struct {
	*Destination
	ReplyToAddresses []string
	ReturnPath       string
	Source           string
	*Message
}

func NewSendEmailRequest() *SendEmailRequest {
	return &SendEmailRequest{
		Destination: new(Destination),
		Message: &Message{
			Body: new(MessageBody),
		},
	}
}
func (req *SendEmailRequest) To(addrs ...string) *SendEmailRequest {
	req.Destination.ToAddresses = append(req.Destination.ToAddresses, addrs...)
	return req
}
func (req *SendEmailRequest) Cc(addrs ...string) *SendEmailRequest {
	req.Destination.CcAddresses = append(req.Destination.CcAddresses, addrs...)
	return req
}
func (req *SendEmailRequest) Bcc(addrs ...string) *SendEmailRequest {
	req.Destination.BccAddresses = append(req.Destination.BccAddresses, addrs...)
	return req
}
func (req *SendEmailRequest) ReplyTo(addrs ...string) *SendEmailRequest {
	req.ReplyToAddresses = append(req.ReplyToAddresses, addrs...)
	return req
}
func (req *SendEmailRequest) Return(path string) *SendEmailRequest {
	req.ReturnPath = path
	return req
}
func (req *SendEmailRequest) Sender(addr string) *SendEmailRequest {
	req.Source = addr
	return req
}
func (req *SendEmailRequest) Subject(data string) *SendEmailRequest {
	req.Message.Subject = &MessageContent{"UTF-8", data}
	return req
}
func (req *SendEmailRequest) Text(data string) *SendEmailRequest {
	req.Message.Body.Text = &MessageContent{"UTF-8", data}
	return req
}
func (req *SendEmailRequest) Html(data string) *SendEmailRequest {
	req.Message.Body.Html = &MessageContent{"UTF-8", data}
	return req
}
func (req SendEmailRequest) Exec(creds aws.Credentials) error {
	dest := req.Destination
	msg := req.Message
	now := time.Now().UTC()

	var numparams = 3 // authentication
	numparams += len(dest.ToAddresses) + len(dest.CcAddresses) + len(dest.BccAddresses)
	numparams += len(req.ReplyToAddresses) + 2
	numparams += 5
	params := make(url.Values, numparams)

	params.Set("AWSAccessKeyId", creds.AccessKeyId)
	params.Set("Action", "SendEmail")
	params.Set("Timestamp", now.Format(time.RFC3339))

	for i := range dest.ToAddresses {
		name := fmt.Sprintf("Destination.ToAddresses.member.%d", i+1)
		params.Set(name, UTF8BEncodedWord(dest.ToAddresses[i]))
	}
	for i := range dest.CcAddresses {
		name := fmt.Sprintf("Destination.CcAddresses.member.%d", i+1)
		params.Set(name, UTF8BEncodedWord(dest.ToAddresses[i]))
	}
	for i := range dest.BccAddresses {
		name := fmt.Sprintf("Destination.BccAddresses.member.%d", i+1)
		params.Set(name, UTF8BEncodedWord(dest.ToAddresses[i]))
	}

	for i := range req.ReplyToAddresses {
		name := fmt.Sprintf("ReplyToAddresses.member.%d", i+1)
		params.Set(name, UTF8BEncodedWord(req.ReplyToAddresses[i]))
	}
	if req.ReturnPath != "" {
		params.Set("ReturnPath", req.ReturnPath)
	}
	params.Set("Source", UTF8BEncodedWord(req.Source))

	params.Set("Message.Subject.Charset", msg.Subject.Charset)
	params.Set("Message.Subject.Data", msg.Subject.Data)
	if msg.Body.Text != nil {
		params.Set("Message.Body.Text.Charset", msg.Body.Text.Charset)
		params.Set("Message.Body.Text.Data", msg.Body.Text.Data)
	}
	if msg.Body.Html != nil {
		params.Set("Message.Body.Html.Charset", msg.Body.Html.Charset)
		params.Set("Message.Body.Html.Data", msg.Body.Html.Data)
	}

	host := "email.us-east-1.amazonaws.com"
	urlstr := "https://" + host + "/SendEmail"
	body := params.Encode()
	httpreq, err := http.NewRequest("POST", urlstr, strings.NewReader(body))
	if err != nil {
		return err
	}

	httpreq.Header.Set("Date", now.Format(time.RFC1123))
	httpreq.Header.Set("Host", host)
	httpreq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpreq.Header.Set("X-Amzn-Authorization",
		fmt.Sprintf("AWS3 AWSAccessKeyID=%s, Signature=%s, Algorithm=%s, SignedHeaders=%s",
			creds.AccessKeyId, "needtodothis", "HmacSHA256", "Date;Host"))

	return nil
}

type Destination struct {
	BccAddresses []string
	CcAddresses  []string
	ToAddresses  []string
}

type Sender struct {
	ReplyToAddresses []string
	ReturnPath       string
	Source           string
}

type Message struct {
	Subject *MessageContent
	Body    *MessageBody
}

func NewMessage(subject, text, html string) *Message {
	m := &Message{Subject: &MessageContent{"UTF-8", subject}}
	m.Body = new(MessageBody)
	if text != "" {
		m.Body.Text = &MessageContent{"UTF-8", text}
	}
	if html != "" {
		m.Body.Html = &MessageContent{"UTF-8", html}
	}
	return m
}

type MessageBody struct {
	Text *MessageContent
	Html *MessageContent
}

type MessageContent struct {
	Charset string
	Data    string
}

func UTF8BEncodedWord(str string) string {
	return fmt.Sprintf("=UTF-8?B?%s=", base64.URLEncoding.EncodeToString([]byte(str)))
}

func UTF8BDecodeWord(enc string) (string, error) {
	firsteq := strings.Index(enc, "=")
	if firsteq != 0 {
		return "", fmt.Errorf("expected prefix \"=\"")
	}
	secondeq := strings.Index(enc[firsteq+1:], "=")
	switch secondeq {
	case -1:
		return "", fmt.Errorf("expected suffix \"=\"")
	case len(enc) - 1:
		break
	default:
		return "", fmt.Errorf("unexpected \"=\" index %d", secondeq)

	}

	pieces := strings.Split(enc[1:len(enc)-1], "?")
	if len(pieces) > 3 {
		return "", fmt.Errorf("unexpected  \"?\" index ???")
	} else if len(pieces) < 3 {
		return "", fmt.Errorf("missing \"?\"")
	}

	charset, encoding, enctext := pieces[0], pieces[1], pieces[2]
	if charset != "UTF-8" && charset != "ISO-8859-1" {
		return "", fmt.Errorf("unsupported charset: %q", charset)
	}
	if encoding != "B" {
		return "", fmt.Errorf("unsupported encoding: %q", encoding)
	}
	p, err := base64.StdEncoding.DecodeString(enctext)

	return string(p), err
}
