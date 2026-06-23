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
		SyncEnabled bool `yaml:"sync_enabled"`
		SyncOnce    bool `yaml:"sync_once"`
		LeagueID    int  `yaml:"league_id"`
		OddsEnabled bool `yaml:"odds_enabled"`
	} `yaml:"fpl"`
	WCF struct {
		SyncEnabled bool `yaml:"sync_enabled"`
		OddsEnabled bool `yaml:"odds_enabled"`
	} `yaml:"wcf"`
	Odds struct {
		CacheTTL string `yaml:"cache_ttl"`
	} `yaml:"odds"`
	Form struct {
		GWWindow int `yaml:"gw_window"`
	} `yaml:"form"`
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

func corsOrigins() []string {
	v := os.Getenv("CORS_ALLOWED_ORIGINS")
	if v != "" {
		return strings.Split(v, ",")
	}
	return []string{"http://localhost:5173"}
}
