package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	RedisURL    string
	Port        string

	FPLSyncEnabled     bool
	FPLSyncIntervalMin int
	FPLLeagueID        int
	FPLTopNDefault     int

	WCFSyncEnabled bool
	WCFAuthToken   string

	UCLFSyncEnabled bool
	UCLFAuthToken   string

	FormGWWindow int

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
		FPLSyncIntervalMin: envInt("FPL_SYNC_INTERVAL_MIN", 30),
		FPLLeagueID:        envInt("FPL_LEAGUE_ID", 314),
		FPLTopNDefault:     envInt("FPL_TOP_N_DEFAULT", 10000),

		WCFSyncEnabled: envBool("WCF_SYNC_ENABLED", true),
		WCFAuthToken:   os.Getenv("WCF_AUTH_TOKEN"),

		UCLFSyncEnabled: envBool("UCLF_SYNC_ENABLED", false),
		UCLFAuthToken:   os.Getenv("UCLF_AUTH_TOKEN"),

		FormGWWindow: envInt("FORM_GW_WINDOW", 3),

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
