package config

import "os"

type Config struct {
	Port                string
	GoogleProjectID     string
	GoogleProjectNumber string
	CollectionName      string
	BucketName          string
	GeminiAPIKey        string
}

func Load() Config {
	return Config{
		Port:                getEnv("PORT", "8080"),
		GoogleProjectID:     getEnv("GOOGLE_PROJECT_ID", "primal-bonbon-464323-d3"),
		GoogleProjectNumber: getEnv("GOOGLE_PROJECT_NUMBER", "837291048792"),
		CollectionName:      getEnv("COLLECTION_NAME", "resume-my-mom-notes-collection"),
		BucketName:          getEnv("BUCKET_NAME", "my-mom-voice-notes"),
		GeminiAPIKey:        getEnv("GEMINI_API_KEY", "GEMINI_API_KEY"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
