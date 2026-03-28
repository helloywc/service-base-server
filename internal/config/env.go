package config

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// LoadEnv 按顺序加载根目录下的 env 文件（后加载文件中的同名键会覆盖先加载文件中的值）。
// 已在进程环境中存在的变量（启动前由 shell/系统注入）一律不覆盖。
//
// 加载顺序：
//  1. .env                    — 全环境通用基线
//  2. .env.{APP_ENV}          — 见 envSpecificFiles()
//  3. .env.local              — 本机覆盖（可选，勿提交）
//
// APP_ENV 常用值：development | test | production（亦接受 dev / prod 别名，仅用于端口与 normalize；环境文件名为 .env.development / .env.test / .env.production）。
func LoadEnv() {
	initial := snapshotEnvKeys()

	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("config: unable to get working directory: %v", err)
		return
	}

	loadEnvFile(filepath.Join(cwd, ".env"), initial)

	appEnv := strings.TrimSpace(os.Getenv("APP_ENV"))
	fileEnv := appEnv
	if fileEnv == "" {
		fileEnv = "development"
	}
	for _, name := range envSpecificFiles(fileEnv) {
		loadEnvFile(filepath.Join(cwd, name), initial)
	}

	loadEnvFile(filepath.Join(cwd, ".env.local"), initial)
}

// envSpecificFiles 返回环境专用文件名（规范命名，按顺序加载；后者覆盖前者）。
func envSpecificFiles(appEnv string) []string {
	switch strings.ToLower(appEnv) {
	case "development", "dev":
		return []string{".env.development"}
	case "test":
		return []string{".env.test"}
	case "production", "prod":
		return []string{".env.production"}
	default:
		if appEnv == "" {
			return nil
		}
		return []string{".env." + appEnv}
	}
}

func snapshotEnvKeys() map[string]struct{} {
	m := make(map[string]struct{})
	for _, e := range os.Environ() {
		if i := strings.IndexByte(e, '='); i > 0 {
			m[e[:i]] = struct{}{}
		}
	}
	return m
}

func loadEnvFile(path string, initial map[string]struct{}) {
	if err := loadEnvFileInner(path, initial); err != nil {
		if os.IsNotExist(err) {
			return
		}
		log.Printf("config: error loading %s: %v", path, err)
	}
}

func loadEnvFileInner(path string, initial map[string]struct{}) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
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
		if _, exists := initial[key]; exists {
			continue
		}
		if err := os.Setenv(key, val); err != nil {
			log.Printf("config: failed to set %s from %s: %v", key, path, err)
		}
	}
	return scanner.Err()
}
