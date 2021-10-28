package gateway

type chatMng struct {
}

var ChatMng *chatMng

func initChatMng() {
	ChatMng = new(chatMng)
}

func (that chatMng) Send(fromId string, toId string, uri string, bytes []byte) (bool, error) {
	fClient := Server.GetProdGid(fromId).GetGWIClient()
	fid, err := fClient.GPush(Server.Context, fromId, uri, bytes, true, 3, false, "", 0)
	if fid < 32 {
		return false, err
	}

	tid, err := Server.GetProdGid(toId).GetGWIClient().GPush(Server.Context, toId, uri, bytes, true, 3, false, "", fid)
	if tid < 32 {
		fClient.GPushA(Server.Context, fromId, fid, false)
		return false, err
	}

	fClient.GPushA(Server.Context, fromId, fid, true)
	return true, err
}
