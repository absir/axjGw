package gateway

import (
	"axj/APro"
	"axj/Kt/Kt"
	"axj/Kt/KtCvt"
	"axj/Kt/KtEncry"
	"axj/Kt/KtUnsafe"
	jsoniter "github.com/json-iterator/go"
)

type uriDict struct {
	uriMapUriI     map[string]int32
	uriIMapUri     map[int32]string
	UriMapJson     string
	UriMapHash     string
	UriMapJsonData []byte
}

func (that *uriDict) UriMapUriI() map[string]int32 {
	return that.uriMapUriI
}

func (that *uriDict) UriIMapUri() map[int32]string {
	return that.uriIMapUri
}

var UriDict *uriDict

func initUriDict() {
	cfg := APro.FileCfg("uriDict.properties")
	if cfg == nil {
		UriDict = new(uriDict)
		UriDict.UriMapHash = ""
		UriDict.uriMapUriI = map[string]int32{}
		UriDict.uriIMapUri = map[int32]string{}

	} else {
		UriDict = new(uriDict)
		UriDict.uriMapUriI = map[string]int32{}
		UriDict.uriIMapUri = map[int32]string{}
		for key, val := range cfg {
			uriI := KtCvt.ToType(key, KtCvt.Int32).(int32)
			uri := KtCvt.ToType(val, KtCvt.String).(string)
			if uri == "" || uriI <= 0 {
				continue
			}

			UriDict.uriMapUriI[uri] = uriI
			UriDict.uriIMapUri[uriI] = uri
		}

		// 排序保证hash一致
		config := jsoniter.Config{
			SortMapKeys: true,
		}

		json, err := config.Froze().MarshalToString(UriDict.uriMapUriI)
		Kt.Panic(err)
		UriDict.UriMapJson = json
		UriDict.UriMapHash = KtEncry.EnMd5(KtUnsafe.StringToBytes(UriDict.UriMapJson))
		UriDict.UriMapJsonData = KtUnsafe.StringToBytes(UriDict.UriMapJson)
	}
}
