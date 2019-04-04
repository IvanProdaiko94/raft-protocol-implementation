package rpc

type Resp struct {
	Val interface{}
	Err error
}

type Call struct {
	In interface{}
	C  chan Resp
}

func (c *Call) Respond(val interface{}, err error) {
	c.C <- Resp{Val: val, Err: err}
}
