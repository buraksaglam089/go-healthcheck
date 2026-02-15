package storage

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/buraksaglam089/go-healthcheck/monitor"
	"github.com/fatih/color"
)

type FileLogger struct {
	FilePath string
}

func NewFileLogger(filePath string) *FileLogger {
	return &FileLogger{
		FilePath: filePath,
	}
}

func (l *FileLogger) SaveLog(result monitor.Result) error {
	jsonData, err := json.Marshal(result)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(l.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	jsonData = append(jsonData, '\n')
	_, err = file.Write(jsonData)
	if err != nil {
		return err
	}
	logResult(result)

	return nil
}

func logResult(res monitor.Result) {
	successColor := color.New(color.FgGreen).SprintfFunc()
	errorColor := color.New(color.FgRed).SprintfFunc()
	warnColor := color.New(color.FgYellow).SprintfFunc()
	timeColor := color.New(color.FgHiBlue).SprintfFunc()

	timestamp := timeColor(res.Timestamp.Format("15:04:05"))

	if res.Err != nil {
		fmt.Printf("%s %s %s %s\n",
			timestamp,
			errorColor("❌"),
			warnColor(res.TargetURL),
			errorColor(res.Err.Error()),
		)
	} else {
		fmt.Printf("%s %s %s %s %s\n",
			timestamp,
			successColor("✅"),
			warnColor(res.TargetURL),
			successColor("Status: %d", res.StatusCode),
			successColor("Latency: %v", res.Latency),
		)
	}
}
