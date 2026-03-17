package config

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// LoadEnv loads environment variables from .env files in a way
// similar to Node's dotenv:
//   - .env
//   - .env.{APP_ENV}
//   - .env.local
// Files that do not exist are skipped. Existing environment
// variables are not overwritten.
func LoadEnv() {
	appEnv := os.Getenv("APP_ENV")

	var files []string
	files = append(files, ".env")
	if appEnv != "" {
		files = append(files, ".env."+appEnv)
		// 开发时常用 .env.dev，与 APP_ENV=development 兼容
		if appEnv == "development" {
			files = append(files, ".env.dev")
		}
	}
	files = append(files, ".env.local")

	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("config: unable to get working directory: %v", err)
		return
	}

	for _, name := range files {
		path := filepath.Join(cwd, name)
		if err := loadEnvFile(path); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			log.Printf("config: error loading %s: %v", path, err)
		}
	}
}

func loadEnvFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines.
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		if key == "" {
			continue
		}

		// Do not overwrite existing environment variables.
		if _, exists := os.LookupEnv(key); !exists {
			if err := os.Setenv(key, val); err != nil {
				log.Printf("config: failed to set %s from %s: %v", key, path, err)
			}
		}
	}

	return scanner.Err()
}

