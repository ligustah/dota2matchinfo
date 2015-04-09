package steam

type Player struct {
	AbilityUpgrades []AbilityUpgrade `msgpack:"ability_upgrades"json:"ability_upgrades"`
	AccountID       uint32           `msgpack:"account_id" json:"account_id"`
	Assists         uint16           `msgpack:"assists" json:"assists"`
	Deaths          uint16           `msgpack:"deaths" json:"deaths"`
	Denies          uint16           `msgpack:"denies" json:"denies"`
	Gold            uint32           `msgpack:"gold" json:"gold"`
	GoldPerMin      uint16           `msgpack:"gold_per_min" json:"gold_per_min"`
	GoldSpent       uint32           `msgpack:"gold_spent" json:"gold_spent"`
	HeroDamage      uint32           `msgpack:"hero_damage" json:"hero_damage"`
	HeroHealing     uint32           `msgpack:"hero_healing" json:"hero_healing"`
	HeroID          uint8            `msgpack:"hero_id" json:"hero_id"`
	Item0           uint8            `msgpack:"item_0" json:"item_0"`
	Item1           uint8            `msgpack:"item_1" json:"item_1"`
	Item2           uint8            `msgpack:"item_2" json:"item_2"`
	Item3           uint8            `msgpack:"item_3" json:"item_3"`
	Item4           uint8            `msgpack:"item_4" json:"item_4"`
	Item5           uint8            `msgpack:"item_5" json:"item_5"`
	Kills           uint16           `msgpack:"kills" json:"kills"`
	LastHits        uint16           `msgpack:"last_hits" json:"last_hits"`
	LeaverStatus    uint8            `msgpack:"leaver_status" json:"leaver_status"`
	Level           uint8            `msgpack:"level" json:"level"`
	PlayerSlot      uint8            `msgpack:"player_slot" json:"player_slot"`
	TowerDamage     uint32           `msgpack:"tower_damage" json:"tower_damage"`
	XpPerMin        uint16           `msgpack:"xp_per_min" json:"xp_per_min"`
}

type AbilityUpgrade struct {
	Ability uint16 `msgpack:"ability" json:"ability"`
	Level   uint8  `msgpack:"level" json:"level"`
	Time    uint16 `msgpack:"time" json:"time"`
}

type Match struct {
	BarracksStatusDire    uint16   `msgpack:"barracks_status_dire" json:"barracks_status_dire"`
	BarracksStatusRadiant uint16   `msgpack:"barracks_status_radiant" json:"barracks_status_radiant"`
	Cluster               uint16   `msgpack:"cluster" json:"cluster"`
	Duration              uint16   `msgpack:"duration" json:"duration"`
	FirstBloodTime        uint16   `msgpack:"first_blood_time" json:"first_blood_time"`
	GameMode              uint8    `msgpack:"game_mode" json:"game_mode"`
	HumanPlayers          uint8    `msgpack:"human_players" json:"human_players"`
	Leagueid              uint16   `msgpack:"leagueid" json:"leagueid"`
	LobbyType             uint8    `msgpack:"lobby_type" json:"lobby_type"`
	MatchID               uint64   `msgpack:"match_id" json:"match_id"`
	MatchSeqNum           uint64   `msgpack:"match_seq_num" json:"match_seq_num"`
	Players               []Player `msgpack:"players" json:"players"`
	NegativeVotes         uint32   `msgpack:"negative_votes" json:"negative_votes"`
	PositiveVotes         uint32   `msgpack:"positive_votes" json:"positive_votes"`
	RadiantWin            bool     `msgpack:"radiant_win" json:"radiant_win"`
	StartTime             uint64   `msgpack:"start_time" json:"start_time"`
	TowerStatusDire       uint16   `msgpack:"tower_status_dire" json:"tower_status_dire"`
	TowerStatusRadiant    uint16   `msgpack:"tower_status_radiant" json:"tower_status_radiant"`
}

type MatchHistoryResult struct {
	Matches          []Match `json:"matches"`
	NumResults       uint    `json:"num_results"`
	ResultsRemaining uint    `json:"results_remaining"`
	Status           uint    `json:"status"`
	TotalResults     uint    `json:"total_results"`
}

type MatchHistoryResponse struct {
	Result *MatchHistoryResult `json:"result"`
}
