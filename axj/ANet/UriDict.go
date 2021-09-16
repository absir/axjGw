package ANet

type UriDict interface {
	UriMapUriI() map[string]int32
	UriIMapUri() map[int32]string
}
