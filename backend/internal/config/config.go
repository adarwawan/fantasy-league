package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	RedisURL    string
	Port        string

	FPLSyncEnabled     bool
	FPLSyncOnce        bool
	FPLSyncIntervalMin int
	FPLLeagueID        int

	WCFSyncEnabled bool
	WCFAuthToken   string

	UCLFSyncEnabled bool
	UCLFAuthToken   string

	FormGWWindow int

	OddsAPIKey     string
	OddsCacheTTL   time.Duration
	WCFOddsEnabled bool
	FPLOddsEnabled bool

	SyncEndpointSecret string
	CORSAllowedOrigins []string
}

func Load() Config {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		RedisURL:    os.Getenv("REDIS_URL"),
		Port:        port,

		FPLSyncEnabled:     envBool("FPL_SYNC_ENABLED", true),
		FPLSyncOnce:        envBool("FPL_SYNC_ONCE", true),
		FPLSyncIntervalMin: envInt("FPL_SYNC_INTERVAL_MIN", 30),
		FPLLeagueID:        envInt("FPL_LEAGUE_ID", 314),

		WCFSyncEnabled: envBool("WCF_SYNC_ENABLED", true),
		WCFAuthToken:   os.Getenv("WCF_AUTH_TOKEN"),

		UCLFSyncEnabled: envBool("UCLF_SYNC_ENABLED", false),
		UCLFAuthToken:   os.Getenv("UCLF_AUTH_TOKEN"),

		FormGWWindow: envInt("FORM_GW_WINDOW", 3),

		OddsAPIKey:     os.Getenv("ODDS_API_KEY"),
		OddsCacheTTL:   envDuration("ODDS_CACHE_TTL", 15*time.Minute),
		WCFOddsEnabled: envBool("WCF_ODDS_ENABLED", false),
		FPLOddsEnabled: envBool("FPL_ODDS_ENABLED", false),

		SyncEndpointSecret: os.Getenv("SYNC_ENDPOINT_SECRET"),
		CORSAllowedOrigins: corsOrigins(),
	}
}

func envBool(key string, def bool) bool {
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

func corsOrigins() []string {
	v := os.Getenv("CORS_ALLOWED_ORIGINS")
	if v != "" {
		return strings.Split(v, ",")
	}
	return []string{"http://localhost:5173"}
}

func envInt(key string, def int) int {
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

func envDuration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}
