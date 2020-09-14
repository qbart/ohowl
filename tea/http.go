package tea

import (
	"context"

	"github.com/imroc/req"
	"gopkg.in/yaml.v2"
)

type HttpReq struct {
	Err  error
	Resp *req.Resp
}

func (r *HttpReq) ToJSON(o interface{}) *HttpReq {
	if r.Err == nil {
		r.Err = r.Resp.ToJSON(o)
	}
	return r
}

func (r *HttpReq) ToYAML(o interface{}) *HttpReq {
	if r.Err == nil {
		r.Err = yaml.Unmarshal(r.Resp.Bytes(), o)
	}
	return r
}

func HttpDelete(ctx context.Context, url string, args ...interface{}) *HttpReq {
	resp, err := req.Delete(url, args...)
	if err != nil {
		return &HttpReq{Err: err, Resp: resp}
	}
	return &HttpReq{Resp: resp}
}

func HttpGet(ctx context.Context, url string, args ...interface{}) *HttpReq {
	resp, err := req.Get(url, args...)
	if err != nil {
		return &HttpReq{Err: err, Resp: resp}
	}
	return &HttpReq{Resp: resp}
}

func HttpPost(ctx context.Context, url string, args ...interface{}) *HttpReq {
	resp, err := req.Post(url, args...)
	if err != nil {
		return &HttpReq{Err: err, Resp: resp}
	}
	return &HttpReq{Resp: resp}
}
