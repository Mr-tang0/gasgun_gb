package services

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func GetConfigPath(company string, software string, subdir string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "config_gasgun2.json"
	}
	DIR := filepath.Join(homeDir, company, software, subdir)
	os.MkdirAll(DIR, 0755)

	return filepath.Join(DIR, "config.json")
}

func GetLogPath(company string, software string, subdir string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "config_gasgun2.json"
	}
	DIR := filepath.Join(homeDir, company, software, subdir, "log")
	os.MkdirAll(DIR, 0755)

	filename := fmt.Sprintf("log_%s.log", time.Now().Format("20060102150405"))

	return filepath.Join(DIR, filename)
}

func Log(msg string, path string) error {
	line := fmt.Sprintf("%s %s\n", time.Now().Format("2006-01-02 15:04:05"), msg)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(line)
	if err != nil {
		return err
	}
	return nil
}
