package common

import (
	"time"
)

const (
	// TIMESTAMP_STR_FORMAT string = "yyyy-mm-ddThh:mm:ss"
	TIMESTAMP_STR_FORMAT       string = time.RFC3339
	GIN_CTX_JWT_CLAIM_KEY_NAME string = "jwtClaims"
	JWT_TIMEOUT_SECS           int64  = 30 * 60
)

var (
	PROJECT_NAME  string = GetEnvVarDefault("PROJECT_NAME", "patos-app")
	NOREPLY_EMAIL string = GetEnvVarDefault("NOREPLY_EMAIL", "no-reply@example.com")
	APP_HOST_URL  string = GetEnvVarDefault("APP_HOST_URL", "http://localhost:8080/")
	API_HOST_URL  string = GetEnvVarDefault("API_HOST_URL", "http://localhost:8080/")
	COOKIE_NAME   string = PROJECT_NAME + "_jwt"
)
