package service

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const databasesBase = "/Users/yang/Operator/Databases"

// Archive 将 name 对应目录打 zip 包，放在同级目录，命名为 name_YYYY-MM-DD_HH-mm-ss.zip
// name "mysql-dev" -> 目录 .../mysql/dev，zip 如 mysql-dev_2026-03-08_06-52-26.zip
func Archive(name string) (zipPath string, err error) {
	// name 中 "-" 替换为 "/" 得到相对路径，如 mysql-dev -> mysql/dev
	rel := strings.ReplaceAll(name, "-", "/")
	dirPath := filepath.Join(databasesBase, rel)
	dirPath = filepath.Clean(dirPath)
	baseClean := filepath.Clean(databasesBase)
	// 防止路径逃逸
	if dirPath != baseClean && !strings.HasPrefix(dirPath, baseClean+string(os.PathSeparator)) {
		return "", fmt.Errorf("invalid path: %s", dirPath)
	}
	info, err := os.Stat(dirPath)
	if err != nil {
		return "", fmt.Errorf("stat folder: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("not a directory: %s", dirPath)
	}

	parentDir := filepath.Dir(dirPath)
	// 文件名中的时间用 YYYY-MM-DD_HH-mm-ss，避免冒号
	ts := time.Now().Format("2006-01-02_15-04-05")
	zipName := name + "_" + ts + ".zip"
	zipPath = filepath.Join(parentDir, zipName)

	fw, err := os.Create(zipPath)
	if err != nil {
		return "", fmt.Errorf("create zip: %w", err)
	}
	defer fw.Close()

	zw := zip.NewWriter(fw)
	defer zw.Close()

	// 压缩目录下的所有文件，zip 内根目录为目录名（如 dev）
	rootInZip := filepath.Base(dirPath) + "/"
	err = filepath.WalkDir(dirPath, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relPath, _ := filepath.Rel(dirPath, path)
		if relPath == "." {
			return nil
		}
		entryName := rootInZip + filepath.ToSlash(relPath)
		if d.IsDir() {
			entryName += "/"
			_, err := zw.Create(entryName)
			return err
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		w, err := zw.Create(entryName)
		if err != nil {
			f.Close()
			return err
		}
		_, err = io.Copy(w, f)
		f.Close()
		return err
	})
	if err != nil {
		os.Remove(zipPath)
		return "", fmt.Errorf("walk dir: %w", err)
	}
	return zipPath, nil
}
