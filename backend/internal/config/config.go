package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	DatabaseURL string
	RedisURL    string
	Port        string

	FPLSyncEnabled bool
	FPLSyncOnce    bool
	FPLLeagueID    int

	WCFSyncEnabled bool
	WCFAuthToken   string

	FormGWWindow int

	// PicksWorkers is the number of managers whose picks are fetched in
	// parallel during a sync.
	PicksWorkers int

	// SyncFreshnessMaxAge is how old a game's last successful sync may be
	// before GET /health/sync reports it unhealthy.
	SyncFreshnessMaxAge time.Duration

	// MustHave holds per-game must-have thresholds, keyed by game ID.
	MustHave map[string]MustHaveConfig

	OddsAPIKey     string
	OddsCacheTTL   time.Duration
	WCFOddsEnabled bool
	FPLOddsEnabled bool

	SyncEndpointSecret string
	CORSAllowedOrigins []string
}

type coldConfig struct {
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
	FPL struct {
		SyncEnabled bool           `yaml:"sync_enabled"`
		SyncOnce    bool           `yaml:"sync_once"`
		LeagueID    int            `yaml:"league_id"`
		OddsEnabled bool           `yaml:"odds_enabled"`
		MustHave    MustHaveConfig `yaml:"must_have"`
	} `yaml:"fpl"`
	WCF struct {
		SyncEnabled bool           `yaml:"sync_enabled"`
		OddsEnabled bool           `yaml:"odds_enabled"`
		MustHave    MustHaveConfig `yaml:"must_have"`
	} `yaml:"wcf"`
	Odds struct {
		CacheTTL string `yaml:"cache_ttl"`
	} `yaml:"odds"`
	Form struct {
		GWWindow int `yaml:"gw_window"`
	} `yaml:"form"`
	Sync struct {
		PicksWorkers    int    `yaml:"picks_workers"`
		FreshnessMaxAge string `yaml:"freshness_max_age"`
	} `yaml:"sync"`
}

// MustHaveConfig holds the thresholds for must-have player detection.
// Each game configures its own set under its yaml section.
type MustHaveConfig struct {
	FormWindow    int     `yaml:"form_window"`
	FormPointsMin int     `yaml:"form_points_min"`
	FormRatio     float64 `yaml:"form_ratio"`
	MaxNextFDR    int     `yaml:"max_next_fdr"`
	TopGK         int     `yaml:"top_gk"`
	TopDEF        int     `yaml:"top_def"`
	TopMID        int     `yaml:"top_mid"`
	TopFWD        int     `yaml:"top_fwd"`
}

func defaultMustHave() MustHaveConfig {
	return MustHaveConfig{
		FormWindow:    5,
		FormPointsMin: 6,
		FormRatio:     0.5,
		MaxNextFDR:    3,
		TopGK:         4,
		TopDEF:        8,
		TopMID:        8,
		TopFWD:        5,
	}
}

func loadColdConfig() coldConfig {
	// defaults — used when config.yaml is missing or a field is absent
	cc := coldConfig{}
	cc.Server.Port = 8080
	cc.FPL.SyncEnabled = true
	cc.FPL.SyncOnce = true
	cc.FPL.LeagueID = 314
	cc.FPL.OddsEnabled = false
	cc.WCF.SyncEnabled = true
	cc.WCF.OddsEnabled = true
	cc.Odds.CacheTTL = "15m"
	cc.Form.GWWindow = 3
	cc.Sync.PicksWorkers = 10
	cc.Sync.FreshnessMaxAge = "26h"
	cc.FPL.MustHave = defaultMustHave()
	cc.WCF.MustHave = defaultMustHave()

	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "config.yaml"
	}
	f, err := os.Open(path)
	if err != nil {
		return cc
	}
	defer f.Close()
	_ = yaml.NewDecoder(f).Decode(&cc)
	return cc
}

func Load() Config {
	_ = godotenv.Load()
	cc := loadColdConfig()

	port := strconv.Itoa(cc.Server.Port)
	if v := os.Getenv("PORT"); v != "" {
		port = v
	}

	oddsCacheTTL, _ := time.ParseDuration(cc.Odds.CacheTTL)
	if oddsCacheTTL == 0 {
		oddsCacheTTL = 15 * time.Minute
	}

	freshnessMaxAge := envDurationOr("SYNC_FRESHNESS_MAX_AGE", cc.Sync.FreshnessMaxAge, 26*time.Hour)

	return Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		RedisURL:    os.Getenv("REDIS_URL"),
		Port:        port,

		FPLSyncEnabled: envBoolOr("FPL_SYNC_ENABLED", cc.FPL.SyncEnabled),
		FPLSyncOnce:    envBoolOr("FPL_SYNC_ONCE", cc.FPL.SyncOnce),
		FPLLeagueID:    envIntOr("FPL_LEAGUE_ID", cc.FPL.LeagueID),

		WCFSyncEnabled: envBoolOr("WCF_SYNC_ENABLED", cc.WCF.SyncEnabled),
		WCFAuthToken:   os.Getenv("WCF_AUTH_TOKEN"),

		FormGWWindow: envIntOr("FORM_GW_WINDOW", cc.Form.GWWindow),

		PicksWorkers:        envIntOr("PICKS_WORKERS", cc.Sync.PicksWorkers),
		SyncFreshnessMaxAge: freshnessMaxAge,

		MustHave: map[string]MustHaveConfig{
			"fpl": cc.FPL.MustHave,
			"wcf": cc.WCF.MustHave,
		},

		OddsAPIKey:     os.Getenv("ODDS_API_KEY"),
		OddsCacheTTL:   oddsCacheTTL,
		WCFOddsEnabled: envBoolOr("WCF_ODDS_ENABLED", cc.WCF.OddsEnabled),
		FPLOddsEnabled: envBoolOr("FPL_ODDS_ENABLED", cc.FPL.OddsEnabled),

		SyncEndpointSecret: os.Getenv("SYNC_ENDPOINT_SECRET"),
		CORSAllowedOrigins: corsOrigins(),
	}
}

func envBoolOr(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

func envIntOr(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}

// envDurationOr resolves a duration from an env var, then a config string,
// falling back to def when neither is set or parseable.
func envDurationOr(key, cfgVal string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	if cfgVal != "" {
		if d, err := time.ParseDuration(cfgVal); err == nil {
			return d
		}
	}
	return def
}

func corsOrigins() []string {
	v := os.Getenv("CORS_ALLOWED_ORIGINS")
	if v != "" {
		return strings.Split(v, ",")
	}
	return []string{"http://localhost:5173"}
}
