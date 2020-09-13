package tea

import "github.com/imroc/req"

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

func HttpDelete(url string, args ...interface{}) *HttpReq {
	resp, err := req.Delete(url, args...)
	if err != nil {
		return &HttpReq{Err: err, Resp: resp}
	}
	return &HttpReq{Resp: resp}
}

func HttpGet(url string, args ...interface{}) *HttpReq {
	resp, err := req.Get(url, args...)
	if err != nil {
		return &HttpReq{Err: err, Resp: resp}
	}
	return &HttpReq{Resp: resp}
}

func HttpPost(url string, args ...interface{}) *HttpReq {
	resp, err := req.Post(url, args...)
	if err != nil {
		return &HttpReq{Err: err, Resp: resp}
	}
	return &HttpReq{Resp: resp}
}
