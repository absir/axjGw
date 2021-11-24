package gateway

import (
	"axj/Kt/Kt"
	"axjGW/gen/gw"
	lru "github.com/hashicorp/golang-lru"
)

type teamMng struct {
	teamMap *lru.Cache
}

var _teamMng *teamMng

func TeamMng() *teamMng {
	if _teamMng == nil {
		Server.Locker.Lock()
		defer Server.Locker.Unlock()
		if _teamMng == nil {
			initTeamMng()
		}
	}

	return _teamMng
}

func initTeamMng() {
	that := new(teamMng)
	var err error
	that.teamMap, err = lru.New(Config.TeamMax)
	Kt.Panic(err)
	_teamMng = that
}

func (that *teamMng) Dirty(tid string) {
	that.teamMap.Remove(tid)
}

func (that *teamMng) GetTeam(tid string) *gw.TeamRep {
	val, _ := that.teamMap.Get(tid)
	team, _ := val.(*gw.TeamRep)
	if team != nil {
		return team
	}

	team, _ = Server.GetProds(Config.AclProd).GetProdHashS(tid).GetAclClient().Team(Server.Context, &gw.GidReq{Gid: tid})
	if team != nil {
		that.teamMap.Add(tid, team)
	}

	return team
}
