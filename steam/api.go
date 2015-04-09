package steam

import (
	"net/http"
	"net/url"
)

const BaseUrl = "api.steampowered.com"

var (
	methodPaths = map[SteamApiMethod]string{
		GetMatchHistory:              "/IDOTA2Match_570/GetMatchHistory/V001/",
		GetMatchDetails:              "/IDOTA2Match_570/GetMatchDetails/v001/",
		GetHeroes:                    "/IEconDOTA2_570/GetHeroes/v0001/",
		GetPlayerSummaries:           "/ISteamUser/GetPlayerSummaries/v0002/",
		EconomySchema:                "/IEconItems_570/GetSchema/v0001/",
		GetLeagueListing:             "/IDOTA2Match_570/GetLeagueListing/v0001/",
		GetLiveLeagueGames:           "/IDOTA2Match_570/GetLiveLeagueGames/v0001/",
		GetMatchHistoryBySequenceNum: "/IDOTA2Match_570/GetMatchHistoryBySequenceNum/v001/",
		GetTeamInfoByTeamID:          "/IDOTA2Match_570/GetTeamInfoByTeamID/v001/",
	}
)

func NewApi(key string) *api {
	return &api{key}
}

type Api interface {
	Request(SteamApiMethod, ...SteamApiArgument) *http.Request
}

type api struct {
	key string
}

func (s *api) Request(method SteamApiMethod, apiArguments ...SteamApiArgument) *http.Request {

	// construct the url query
	arguments := make(url.Values)
	for _, arg := range apiArguments {
		arg(&arguments)
	}

	// set the api key parameter
	arguments.Set("key", s.key)

	u := new(url.URL)
	u.Host = BaseUrl
	u.Path = methodPaths[method]
	u.RawQuery = arguments.Encode()
	u.Scheme = "https"

	// build the request
	req := &http.Request{
		Method:     "GET",
		URL:        u,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Host:       u.Host,
	}

	return req
}

//go:generate stringer -type=SteamApiMethod
type SteamApiMethod int

const (
	GetMatchHistory SteamApiMethod = iota
	GetMatchDetails
	GetHeroes
	GetPlayerSummaries
	EconomySchema
	GetLeagueListing
	GetLiveLeagueGames
	GetMatchHistoryBySequenceNum
	GetTeamInfoByTeamID
)

type steamApiRequest struct {
	method    SteamApiMethod
	arguments url.Values
}
