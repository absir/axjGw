package gateway

import (
	"axj/Kt/Kt"
	"axjGW/gen/gw"
	lru "github.com/hashicorp/golang-lru"
)

type teamMng struct {
	teamMap *lru.Cache
}

var TeamMng *teamMng

func initTeamMng() {
	TeamMng = new(teamMng)
	var err error
	TeamMng.teamMap, err = lru.New(Config.TeamMax)
	Kt.Panic(err)
}

func (that *teamMng) Dirty(tid string) {
	that.teamMap.Remove(tid)
}

func (that *teamMng) GetTeam(tid string) *gw.Team {
	val, _ := that.teamMap.Get(tid)
	team := val.(*gw.Team)
	if team != nil {
		return team
	}

	team, _ = Server.GetProds(Config.AclProd).GetProdHashS(tid).GetAclClient().Team(Server.Context, tid)
	if team != nil {
		that.teamMap.Add(tid, team)
	}

	return team
}
