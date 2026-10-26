package filesystem

import (
	"os"
	"path/filepath"
	"strings"
)

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// isTempDir: go run собирает exe во временную папку
func isTempDir(dir string) bool {
	low := strings.ToLower(dir)
	return strings.Contains(low, "go-build") ||
		strings.HasPrefix(low, strings.ToLower(os.TempDir()))
}

// BaseDir возвращает папку, где лежит (или будет создан) конфиг
func BaseDir(configFile string) string {

	wd, wdErr := os.Getwd()
	exe, exeErr := os.Executable()
	exeDir := filepath.Dir(exe)

	// Если запущено из IDE / go run: конфиг лежит в рабочей папке
	if wdErr == nil && fileExists(filepath.Join(wd, configFile)) {
		return wd
	}

	// Если запущен готовый exe: конфиг лежит рядом с ним
	if exeErr == nil && fileExists(filepath.Join(exeDir, configFile)) {
		return exeDir
	}

	// Конфига ещё нет. При go run exe во временной папке, создаём конфиг в рабочей
	if wdErr == nil && (exeErr != nil || isTempDir(exeDir)) {
		return wd
	}

	if exeErr == nil {
		return exeDir
	}

	return "."
}
