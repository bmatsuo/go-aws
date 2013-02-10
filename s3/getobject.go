package s3

import (
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/bmatsuo/go-aws"
)

type GetObject struct {
	base   baseRequest
	bucket string
	client *Client
}
type GetObjectResponse struct {
	resp   *http.Response
	Header http.Header
	Body   io.ReadCloser
}

func (response *GetObjectResponse) Status() string {
	return response.resp.Status
}
func (response *GetObjectResponse) StatusCode() int {
	return response.resp.StatusCode
}

func (client *Client) GetObject(bucket, key string) *GetObject {
	return &GetObject{
		base: baseRequest{
			Method: "GET",
			Path:   "/" + bucket + "/" + key,
			Query:  make(url.Values, 1),
			Header: make(http.Header, 3), // must not be nil
		},
		bucket: bucket,
		client: client,
	}
}

func (request *GetObject) Exec() (*GetObjectResponse, error) {
	resp, err := request.client.Do(request)
	if err != nil {
		return nil, err
	}
	response := &GetObjectResponse{
		resp:   resp,
		Header: resp.Header,
		Body:   resp.Body,
	}
	return response, nil
}

func (request *GetObject) Request(region *aws.Region) (*http.Request, error) {
	uri := region.Url("", request.base.Path, nil)
	req, err := http.NewRequest(request.base.Method, uri.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header = request.base.Header
	return req, nil
}

func (request *GetObject) ResponseContentType(mime string) *GetObject {
	request.base.Query.Add("response-content-type", mime)
	return request
}
func (request *GetObject) ResponseContentLanguage(q string) *GetObject {
	request.base.Query.Add("response-content-language", q)
	return request
}
func (request *GetObject) ResponseExpires(q string) *GetObject {
	// TODO convert from time.Time
	request.base.Query.Add("response-expires", q)
	return request
}
func (request *GetObject) ResponseResponseCacheControl(q string) *GetObject {
	// TODO convert
	request.base.Query.Add("response-cache-control", q)
	return request
}
func (request *GetObject) ResponseContentDisposition(q string) *GetObject {
	request.base.Query.Add("response-content-disposition", q)
	return request
}
func (request *GetObject) ResponseContentEncoding(q string) *GetObject {
	request.base.Query.Add("response-content-encoding", q)
	return request
}

func (request *GetObject) Range(h string) *GetObject {
	// TODO convert
	request.base.Header.Add("Range", h)
	return request
}
func (request *GetObject) IfModifiedSince(latest time.Time) *GetObject {
	request.base.Header.Add("If-Modified-Since", latest.UTC().Format(time.RFC1123))
	return request
}
func (request *GetObject) IfUnmodifiedSince(latest time.Time) *GetObject {
	request.base.Header.Add("If-Unmodified-Since", latest.UTC().Format(time.RFC1123))
	return request
}
func (request *GetObject) IfMatch(etag string) *GetObject {
	request.base.Header.Add("If-Match", etag)
	return request
}
func (request *GetObject) IfNoneMatch(etag string) *GetObject {
	request.base.Header.Add("If-None-Match", etag)
	return request
}
