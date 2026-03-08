package service

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const databasesBase = "/Users/yang/Operator/Databases"

// parentDirForName 根据 name 解析目录路径并返回其父目录（与 Archive 中 zip 所在目录一致）
func parentDirForName(name string) (string, error) {
	rel := strings.ReplaceAll(name, "-", "/")
	dirPath := filepath.Join(databasesBase, rel)
	dirPath = filepath.Clean(dirPath)
	baseClean := filepath.Clean(databasesBase)
	if dirPath != baseClean && !strings.HasPrefix(dirPath, baseClean+string(os.PathSeparator)) {
		return "", fmt.Errorf("invalid path: %s", dirPath)
	}
	return filepath.Dir(dirPath), nil
}

// 文件名中的时间格式与 zip 命名一致：name_2006-01-02_15-04-05.zip
const archiveTimeLayout = "2006-01-02_15-04-05"

// ListArchiveFiles 列出以 name_ 或 name- 开头的所有文件，仅返回文件名（无路径、无 .zip 后缀），按日期倒序（越新的越前）
func ListArchiveFiles(name string) ([]string, error) {
	parentDir, err := parentDirForName(name)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(parentDir)
	if err != nil {
		return nil, fmt.Errorf("read dir: %w", err)
	}
	prefixUnderscore := name + "_"
	prefixHyphen := name + "-"
	type fileWithTime struct {
		displayName string
		t           time.Time
	}
	var list []fileWithTime
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		var suffix string
		if strings.HasPrefix(e.Name(), prefixUnderscore) {
			suffix = strings.TrimPrefix(e.Name(), prefixUnderscore)
		} else if strings.HasPrefix(e.Name(), prefixHyphen) {
			suffix = strings.TrimPrefix(e.Name(), prefixHyphen)
		} else {
			continue
		}
		suffix = strings.TrimSuffix(suffix, ".zip")
		t, err := time.Parse(archiveTimeLayout, suffix)
		if err != nil {
			t = time.Time{} // 解析失败排到最末
		}
		displayName := strings.TrimSuffix(e.Name(), ".zip")
		list = append(list, fileWithTime{displayName, t})
	}
	// 按时间倒序：越新越前
	sort.Slice(list, func(i, j int) bool { return list[i].t.After(list[j].t) })
	out := make([]string, len(list))
	for i, f := range list {
		out[i] = f.displayName
	}
	return out, nil
}

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

// Extract 将 name 对应的 zip（timestamp 如 2026-03-08_07-16-48）解压到该 zip 所在目录（当前目录）
func Extract(name, timestamp string) error {
	parentDir, err := parentDirForName(name)
	if err != nil {
		return err
	}
	zipName1 := name + "_" + timestamp + ".zip"
	zipName2 := name + "-" + timestamp + ".zip"
	var zipPath string
	for _, n := range []string{zipName1, zipName2} {
		p := filepath.Join(parentDir, n)
		if _, err := os.Stat(p); err == nil {
			zipPath = p
			break
		}
	}
	if zipPath == "" {
		return fmt.Errorf("zip not found: %s or %s", zipName1, zipName2)
	}
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}
	defer r.Close()
	destDir := filepath.Clean(parentDir)
	for _, f := range r.File {
		dest := filepath.Join(destDir, f.Name)
		cleanDest := filepath.Clean(dest)
		if cleanDest != destDir && !strings.HasPrefix(cleanDest, destDir+string(os.PathSeparator)) {
			return fmt.Errorf("zip slip: %s", f.Name)
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(dest, 0755); err != nil {
				return fmt.Errorf("mkdir %s: %w", dest, err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return fmt.Errorf("mkdir %s: %w", filepath.Dir(dest), err)
		}
		out, err := os.Create(dest)
		if err != nil {
			return fmt.Errorf("create %s: %w", dest, err)
		}
		rc, err := f.Open()
		if err != nil {
			out.Close()
			return fmt.Errorf("open zip entry %s: %w", f.Name, err)
		}
		_, err = io.Copy(out, rc)
		rc.Close()
		out.Close()
		if err != nil {
			return fmt.Errorf("write %s: %w", dest, err)
		}
	}
	return nil
}

// 从文件名（无 .zip）解析 name，支持 name_timestamp 或 name-timestamp
var archiveFilenameRegex = regexp.MustCompile(`^(.+)[_-](\d{4}-\d{2}-\d{2}_\d{2}-\d{2}-\d{2})$`)

// DeleteArchiveFiles 根据文件名数组删除对应 zip，文件名无 .zip 后缀；返回已删除列表与失败列表（含原因）
func DeleteArchiveFiles(filenames []string) (deleted, failed []string) {
	for _, filename := range filenames {
		filename = strings.TrimSpace(filename)
		if filename == "" {
			continue
		}
		// 支持 name_timestamp 或 name-timestamp（后者正则只匹配 _，需兼容）
		subs := archiveFilenameRegex.FindStringSubmatch(filename)
		if len(subs) != 3 {
			failed = append(failed, filename+": invalid format (need name_YYYY-MM-DD_HH-mm-ss)")
			continue
		}
		name := subs[1]
		parentDir, err := parentDirForName(name)
		if err != nil {
			failed = append(failed, filename+": "+err.Error())
			continue
		}
		// 只允许 name_ 或 name- 开头的文件名
		if !strings.HasPrefix(filename, name+"_") && !strings.HasPrefix(filename, name+"-") {
			failed = append(failed, filename+": name mismatch")
			continue
		}
		zipPath := filepath.Join(parentDir, filename+".zip")
		zipPath = filepath.Clean(zipPath)
		if zipPath != parentDir && !strings.HasPrefix(zipPath, parentDir+string(os.PathSeparator)) {
			failed = append(failed, filename+": path invalid")
			continue
		}
		if err := os.Remove(zipPath); err != nil {
			failed = append(failed, filename+": "+err.Error())
			continue
		}
		deleted = append(deleted, filename)
	}
	return deleted, failed
}
