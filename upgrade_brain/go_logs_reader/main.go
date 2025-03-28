package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

type LogEntry struct {
	Timestamp    string `json:"@timestamp"`
	Message      string `json:"message"`
	LoggerName   string `json:"logger_name"`
	ThreadName   string `json:"thread_name"`
	Level        string `json:"level"`
	CustomerID   string `json:"customerId"`
	DataMart     string `json:"datamartMnemonics"`
	SubRequestID string `json:"subRequestId"`
	RequestID    string `json:"requestId"`
}

func main() {
	inputFile := "/var/log/datamart/datamart-3/podd-agent-59.log"
	outputFile := "filtered_logs.log"

	startTime, _ := time.Parse(time.RFC3339, "2024-11-06T00:30:00Z")
	endTime, _ := time.Parse(time.RFC3339, "2024-11-06T02:30:00Z")

	// Читаем файл
	data, err := ioutil.ReadFile(inputFile)
	if err != nil {
		fmt.Println("Ошибка чтения файла:", err)
		return
	}

	var filteredLogs []LogEntry

	// Парсим JSON
	lines := string(data)
	for _, line := range splitLines(lines) {
		var log LogEntry
		if err := json.Unmarshal([]byte(line), &log); err == nil {
			logTime, _ := time.Parse(time.RFC3339, log.Timestamp)
			if logTime.After(startTime) && logTime.Before(endTime) {
				filteredLogs = append(filteredLogs, log)
			}
		}
	}

	// Записываем отфильтрованные логи в выходной файл
	outputFileHandler, err := os.Create(outputFile)
	if err != nil {
		fmt.Println("Ошибка создания выходного файла:", err)
		return
	}
	defer outputFileHandler.Close()

	for _, log := range filteredLogs {
		jsonLine, _ := json.Marshal(log)
		outputFileHandler.Write(jsonLine)
		outputFileHandler.Write([]byte("\n"))
	}

	fmt.Printf("Filtered logs saved to %s\n", outputFile)
}

// splitLines разделяет текст на строки
func splitLines(s string) []string {
	return strings.Split(s, "\n")
}
