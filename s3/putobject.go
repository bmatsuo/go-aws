package s3

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/bmatsuo/go-aws"
)

type PutObject struct {
	base   baseRequest
	bucket string
	body   io.Reader
	client *Client
}
type PutObjectResponse struct {
	resp   *http.Response
	Header http.Header
}

func (response *PutObjectResponse) Status() string {
	return response.resp.Status
}
func (response *PutObjectResponse) StatusCode() int {
	return response.resp.StatusCode
}

func (client *Client) PutObject(bucket, key string) *PutObject {
	return &PutObject{
		base: baseRequest{
			Method: "PUT",
			Path:   "/" + bucket + "/" + key,
			Header: make(http.Header, 3), // must not be nil
		},
		bucket: bucket,
		body:   nil,
		client: client,
	}
}
func (request *PutObject) Exec() (*PutObjectResponse, error) {
	resp, err := request.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	response := &PutObjectResponse{
		resp:   resp,
		Header: resp.Header,
	}
	return response, nil
}

func (request *PutObject) Request(region *aws.Region) (*http.Request, error) {
	uri := region.Url("", request.base.Path, nil)
	req, err := http.NewRequest(request.base.Method, uri.String(), request.body)
	if err != nil {
		return nil, err
	}
	req.Header = request.base.Header
	return req, nil
}

func (request *PutObject) Content(data []byte) *PutObject {
	request.body = bytes.NewBuffer(data)
	request.base.Header.Add("Content-Length", strconv.Itoa(len(data)))
	h := md5.New()
	h.Write(data)
	sum := base64.StdEncoding.EncodeToString(h.Sum(nil))
	request.base.Header.Add("Content-MD5", sum)
	return request
}
func (request *PutObject) CopySource(bucket, key string) *PutObject {
	request.base.Header.Add("x-amz-copy-source", bucket+"/"+key)
	return request
}
func (request *PutObject) MetadataDirective(copyreplace string) *PutObject {
	request.base.Header.Add("x-amz-metadata-directive", copyreplace)
	return request
}
func (request *PutObject) CopySourceIfMatch(etag string) *PutObject {
	request.base.Header.Add("x-amz-copy-source-if-match", etag)
	return request
}
func (request *PutObject) CopySourceIfNoneMatch(etag string) *PutObject {
	request.base.Header.Add("x-amz-copy-source-if-none-match", etag)
	return request
}
func (request *PutObject) CopySourceIfUnmodifiedSince(latest time.Time) *PutObject {
	request.base.Header.Add("x-amz-copy-source-if-unmodified-since", latest.UTC().Format(time.RFC1123))
	return request
}
func (request *PutObject) CopySourceIfModifiedSince(latest time.Time) *PutObject {
	request.base.Header.Add("x-amz-copy-source-if-modified-since", latest.UTC().Format(time.RFC1123))
	return request
}
func (request *PutObject) ContentEncoding(enc string) *PutObject {
	request.base.Header.Add("Content-Encoding", enc)
	return request
}
func (request *PutObject) ContentType(mime string) *PutObject {
	request.base.Header.Add("Content-Type", mime)
	return request
}
func (request *PutObject) CacheControl(control string) *PutObject {
	// TODO marshal
	request.base.Header.Add("Cache-Control", control)
	return request
}
func (request *PutObject) Expires(lifetime time.Duration) *PutObject {
	ms := int64(lifetime)/1e3
	request.base.Header.Add("Expires", strconv.FormatInt(ms, 10))
	return request
}
func (request *PutObject) ServerSideEncryption(algorithm string) *PutObject {
	request.base.Header.Add("x-amz-server-side-encription", algorithm)
	return request
}
func (request *PutObject) StorageClass(class string) *PutObject {
	request.base.Header.Add("x-amz-storage-class", class)
	return request
}
func (request *PutObject) WebsiteRedirectLocation(uri string) *PutObject {
	request.base.Header.Add("x-amz-website-redirect-location", uri)
	return request
}
func (request *PutObject) Acl(acl string) *PutObject {
	request.base.Header.Add("x-amz-acl", acl)
	return request
}
func (request *PutObject) Read(grantee AclGrantee) *PutObject {
	request.base.Header.Add("x-amz-grant-read", grantee.String())
	return request
}
func (request *PutObject) write(grantee AclGrantee) *PutObject {
	request.base.Header.Add("x-amz-grant-write", grantee.String())
	return request
}
func (request *PutObject) ReadAcp(grantee AclGrantee) *PutObject {
	request.base.Header.Add("x-amz-grant-read-acp", grantee.String())
	return request
}
func (request *PutObject) WriteAcp(grantee AclGrantee) *PutObject {
	request.base.Header.Add("x-amz-grant-write-acp", grantee.String())
	return request
}
func (request *PutObject) FullControl(grantee AclGrantee) *PutObject {
	request.base.Header.Add("x-amz-grant-full-control", grantee.String())
	return request
}

type AclGrantee struct {
	EmailAddress string
	Id           string
	Uri          string
}

func (grantee AclGrantee) String() string {
	switch {
	case grantee.EmailAddress != "":
		return fmt.Sprintf("emailAddress=%q", grantee.EmailAddress)
	case grantee.Id != "":
		return fmt.Sprintf("id=%q", grantee.Id)
	case grantee.Uri != "":
		return fmt.Sprintf("uri=%q", grantee.Uri)
	}
	return ""
}
