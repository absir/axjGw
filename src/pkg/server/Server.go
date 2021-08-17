package server

type Client interface {
	Send(uri string, data []byte) bool
}
