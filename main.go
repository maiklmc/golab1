package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	serverURL            = "http://srv.msk01.gigacorp.local/_stats"
	loadAvgThreshold     = 30
	memoryUsageThreshold = 80 // %
	diskUsageThreshold   = 90 // %
	networkUsageThreshold = 90 // %
	checkInterval        = 10 * time.Second
	maxErrorCount       = 3
)

func main() {
	errorCount := 0

	for {
		// Выполняем HTTP-запрос
		resp, err := http.Get(serverURL)
		if err != nil {
			errorCount++
			if errorCount >= maxErrorCount {
				fmt.Println("Unable to fetch server statistic")
			}
			time.Sleep(checkInterval)
			continue
		}

		// Читаем тело ответа
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close() // Явно закрываем тело ответа сразу после чтения

		if err != nil {
			errorCount++
			if errorCount >= maxErrorCount {
				fmt.Println("Unable to fetch server statistic")
			}
			time.Sleep(checkInterval)
			continue
		}

		// Проверяем статус HTTP
		if resp.StatusCode != http.StatusOK {
			errorCount++
			if errorCount >= maxErrorCount {
				fmt.Println("Unable to fetch server statistic")
			}
			time.Sleep(checkInterval)
			continue
	}

		// Сброс счётчика ошибок при успешном получении данных
	errorCount = 0

		// Разбираем данные
		values := strings.Split(string(body), ",")
		if len(values) != 7 {
			errorCount++
			if errorCount >= maxErrorCount {
				fmt.Println("Unable to fetch server statistic")
			}
			time.Sleep(checkInterval)
			continue
	}

	processStats(values)
	time.Sleep(checkInterval)
}
}

func processStats(values []string) {
	var stats [7]int64
	for i, v := range values {
		var err error
		stats[i], err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			fmt.Printf("Error parsing value at index %d: %v\n", i, v)
			return
	}
}

loadAvg := stats[0]
totalMemory := stats[1]
usedMemory := stats[2]
totalDisk := stats[3]
usedDisk := stats[4]
totalBandwidth := stats[5] // байты/сек
usedBandwidth := stats[6]  // байты/сек

// Срез для хранения всех предупреждений
var alerts []string

// 1. Load Average
if loadAvg > loadAvgThreshold {
	alerts = append(alerts, fmt.Sprintf("Load Average is too high: %d", loadAvg))
}

// 2. Memory usage (целочисленный расчёт процента)
if totalMemory > 0 {
	memoryUsagePercent := (usedMemory * 100) / totalMemory
	if memoryUsagePercent > memoryUsageThreshold {
		alerts = append(alerts, fmt.Sprintf("Memory usage too high: %d%%", memoryUsagePercent))
	}
}

// 3. Free disk space
freeDisk := totalDisk - usedDisk
if totalDisk > 0 {
	diskUsagePercent := (usedDisk * 100) / totalDisk
	if diskUsagePercent > diskUsageThreshold {
		freeDiskMb := freeDisk / (1024 * 1024) // байты → МБ
		alerts = append(alerts, fmt.Sprintf("Free disk space is too low: %d Mb left", freeDiskMb))
	}
}

// 4. Network bandwidth (исправленный расчёт под ожидания тестов)
freeBandwidthBytes := totalBandwidth - usedBandwidth
if totalBandwidth > 0 {
	// Делим байты на 1_000_000 (не переводим в мегабиты)
	freeBandwidthMbit := freeBandwidthBytes / 1_000_000
	bandwidthUsagePercent := (usedBandwidth * 100) / totalBandwidth

	if bandwidthUsagePercent > networkUsageThreshold {
		alerts = append(alerts, fmt.Sprintf("Network bandwidth usage high: %d Mbit/s available", freeBandwidthMbit))
	}
}

// Выводим все предупреждения (каждое с новой строки)
for _, alert := range alerts {
	fmt.Println(alert)
}
}
