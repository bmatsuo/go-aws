package s3

import (
	"io"
	"net/http"

	"github.com/bmatsuo/go-aws"
)

type DeleteObject struct {
	base   baseRequest
	bucket string
	client *Client
}
type DeleteObjectResponse struct {
	resp   *http.Response
	Header http.Header
	Body   io.ReadCloser
}

func (response *DeleteObjectResponse) Status() string {
	return response.resp.Status
}
func (response *DeleteObjectResponse) StatusCode() int {
	return response.resp.StatusCode
}

func (client *Client) DeleteObject(bucket, key string) *DeleteObject {
	return &DeleteObject{
		base: baseRequest{
			Method: "DELETE",
			Path:   "/" + bucket + "/" + key,
			Header: make(http.Header, 3), // must not be nil
		},
		bucket: bucket,
		client: client,
	}
}

func (request *DeleteObject) Exec() (*DeleteObjectResponse, error) {
	resp, err := request.client.Do(request)
	if err != nil {
		return nil, err
	}
	response := &DeleteObjectResponse{
		resp:   resp,
		Header: resp.Header,
		Body:   resp.Body,
	}
	return response, nil
}

func (request *DeleteObject) Request(region *aws.Region) (*http.Request, error) {
	uri := region.Url("", request.base.Path, nil)
	req, err := http.NewRequest(request.base.Method, uri.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header = request.base.Header
	return req, nil
}

func (request *DeleteObject) MFA(serial, value string) *DeleteObject {
	request.base.Header.Add("x-amz-mfa", serial+" "+value)
	return request
}
