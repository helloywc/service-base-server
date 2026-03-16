package config

import (
	"bufio"
	"os"
	"strings"
)

// LoadEnv 根据 APP_ENV 直接加载 .env.dev 或 .env.prod（无其它逻辑）
// APP_ENV=prod 时加载 .env.prod，否则加载 .env.dev
// 需在 main 入口最前面调用；若未设 APP_ENV，则加载 .env.dev
func LoadEnv() {
	if os.Getenv("APP_ENV") == "prod" {
		loadFile(".env.prod")
		return
	}
	loadFile(".env.dev")
}

func loadFile(name string) {
	f, err := os.Open(name)
	if err != nil {
		return
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		i := strings.Index(line, "=")
		if i <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:i])
		val := strings.TrimSpace(line[i+1:])
		val = strings.Trim(val, "\"'")
		if key != "" {
			os.Setenv(key, val)
		}
	}
}
