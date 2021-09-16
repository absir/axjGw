package ANet

type UriDict interface {
	UriMapUriI() map[string]int
	UriIMapUri() map[int]string
}
