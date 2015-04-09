package steam

import (
	"net/url"
	"strconv"
)

func RawArgument(key, value string) SteamApiArgument {
	return func(values *url.Values) {
		values.Add(key, value)
	}
}

func RawIntArgument(key string, value int) SteamApiArgument {
	return RawArgument(key, strconv.Itoa(value))
}

func HeroId(id int) SteamApiArgument {
	return RawIntArgument("hero_id", id)
}

func GameMode(id int) SteamApiArgument {
	return RawIntArgument("game_mode", id)
}

func Skill(id int) SteamApiArgument {
	return RawIntArgument("skill", id)
}

func MinPlayers(num int) SteamApiArgument {
	return RawIntArgument("min_players", num)
}

func AccountId(id int) SteamApiArgument {
	return RawIntArgument("account_id", id)
}

func LeagueId(id int) SteamApiArgument {
	return RawIntArgument("league_id", id)
}

func StartAtMatchId(id uint64) SteamApiArgument {
	return RawIntArgument("start_at_match_id", int(id))
}

func StartAtMatchSeqNum(id uint64) SteamApiArgument {
	return RawIntArgument("start_at_match_seq_num", int(id))
}

func MatchesRequested(num int) SteamApiArgument {
	return RawIntArgument("matches_requested", num)
}

func TournamentGamesOnly() SteamApiArgument {
	return RawArgument("tournament_games_only", "true")
}

type SteamApiArgument func(*url.Values)
