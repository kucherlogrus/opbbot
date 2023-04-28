package raiderio

import "time"

type AutoGeneratedErrorMessage struct {
	StatusCode int    `json:"statusCode"`
	Error      string `json:"error"`
	Message    string `json:"message"`
}

type AutoGeneratedActiveAffixes struct {
	Region         string `json:"region"`
	Title          string `json:"title"`
	LeaderboardURL string `json:"leaderboard_url"`
	AffixDetails   []struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		WowheadURL  string `json:"wowhead_url"`
	} `json:"affix_details"`
}

type AutoGeneratedPlayerInfo struct {
	Name                     string    `json:"name"`
	Race                     string    `json:"race"`
	Class                    string    `json:"class"`
	ActiveSpecName           string    `json:"active_spec_name"`
	ActiveSpecRole           string    `json:"active_spec_role"`
	Gender                   string    `json:"gender"`
	Faction                  string    `json:"faction"`
	ThumbnailURL             string    `json:"thumbnail_url"`
	Region                   string    `json:"region"`
	Realm                    string    `json:"realm"`
	LastCrawledAt            time.Time `json:"last_crawled_at"`
	ProfileURL               string    `json:"profile_url"`
	ProfileBanner            string    `json:"profile_banner"`
	MythicPlusScoresBySeason []struct {
		Season string `json:"season"`
		Scores struct {
			All float64 `json:"all"`
		} `json:"scores"`
	} `json:"mythic_plus_scores_by_season"`
	MythicPlusRanks struct {
		Overall struct {
			World  int `json:"world"`
			Region int `json:"region"`
			Realm  int `json:"realm"`
		} `json:"overall"`
		Class struct {
			World  int `json:"world"`
			Region int `json:"region"`
			Realm  int `json:"realm"`
		} `json:"class"`
		FactionOverall struct {
			World  int `json:"world"`
			Region int `json:"region"`
			Realm  int `json:"realm"`
		} `json:"faction_overall"`
		FactionClass struct {
			World  int `json:"world"`
			Region int `json:"region"`
			Realm  int `json:"realm"`
		} `json:"faction_class"`
	} `json:"mythic_plus_ranks"`
	MythicPlusBestRuns []struct {
		Dungeon             string    `json:"dungeon"`
		ShortName           string    `json:"short_name"`
		MythicLevel         int       `json:"mythic_level"`
		CompletedAt         time.Time `json:"completed_at"`
		ClearTimeMs         int       `json:"clear_time_ms"`
		ParTimeMs           int       `json:"par_time_ms"`
		NumKeystoneUpgrades int       `json:"num_keystone_upgrades"`
		MapChallengeModeID  int       `json:"map_challenge_mode_id"`
		ZoneID              int       `json:"zone_id"`
		Score               float64   `json:"score"`
		Affixes             []struct {
			ID          int    `json:"id"`
			Name        string `json:"name"`
			Description string `json:"description"`
			Icon        string `json:"icon"`
			WowheadURL  string `json:"wowhead_url"`
		} `json:"affixes"`
		URL string `json:"url"`
	} `json:"mythic_plus_best_runs"`
	Gear struct {
		UpdatedAt         time.Time `json:"updated_at"`
		ItemLevelEquipped int       `json:"item_level_equipped"`
		ItemLevelTotal    int       `json:"item_level_total"`
	} `json:"gear"`
}
