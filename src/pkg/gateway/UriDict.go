package gateway

import (
	"axj/APro"
	"axj/Kt/KtCvt"
	"axj/Kt/KtEncry"
	"axj/Kt/KtJson"
	"axj/Kt/KtUnsafe"
)

type uriDict struct {
	uriMapUriI     map[string]int32
	uriIMapUri     map[int32]string
	UriMapJson     string
	UriMapHash     string
	UriMapJsonData []byte
}

func (u uriDict) UriMapUriI() map[string]int32 {
	return u.uriMapUriI
}

func (u uriDict) UriIMapUri() map[int32]string {
	return u.uriIMapUri
}

var UriDict *uriDict

func init() {
	cfg := APro.FileCfg("uriDict.properties")
	if cfg == nil {
		UriDict = nil

	} else {
		UriDict = new(uriDict)
		UriDict.uriMapUriI = map[string]int32{}
		UriDict.uriIMapUri = map[int32]string{}
		for key, val := range cfg {
			uri := KtCvt.ToType(key, KtCvt.String).(string)
			uriI := KtCvt.ToType(val, KtCvt.Int32).(int32)
			if uri == "" || uriI <= 0 {
				continue
			}

			UriDict.uriMapUriI[uri] = uriI
			UriDict.uriIMapUri[uriI] = uri
		}

		UriDict.UriMapJson, _ = KtJson.ToJsonStr(UriDict.uriMapUriI)
		UriDict.UriMapHash = KtEncry.EnMd5(KtUnsafe.StringToBytes(UriDict.UriMapJson))
		UriDict.UriMapJsonData = KtUnsafe.StringToBytes(UriDict.UriMapJson)
	}
}
