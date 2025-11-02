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
	serverURL             = "http://srv.msk01.gigacorp.local/_stats"
	loadAvgThreshold      = 30.0
	memoryUsageThreshold  = 80.0 // %
	diskUsageThreshold    = 90.0 // %
	networkUsageThreshold = 90.0 // %
	checkInterval         = 10 * time.Second
	maxErrorCount         = 3
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
	var stats [7]float64
	for i, v := range values {
		var err error
		stats[i], err = strconv.ParseFloat(v, 64)
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

	// 1. Load Average
	if loadAvg > loadAvgThreshold {
		fmt.Printf("Load Average is too high: %.0f\n", loadAvg)
	}

	// 2. Memory usage (с округлением до целого процента)
	if totalMemory > 0 {
		memoryUsagePercent := int((usedMemory / totalMemory) * 100)
		if memoryUsagePercent > memoryUsageThreshold {
			fmt.Printf("Memory usage too high: %d%%\n", memoryUsagePercent)
		}
	}

	// 3. Free disk space
	freeDisk := totalDisk - usedDisk
	if totalDisk > 0 {
		diskUsagePercent := (usedDisk / totalDisk) * 100
		if diskUsagePercent > diskUsageThreshold {
			freeDiskMb := int64(freeDisk) / (1024 * 1024) // байты → МБ
			fmt.Printf("Free disk space is too low: %d Mb left\n", freeDiskMb)
		}
	}

	// 4. Network bandwidth (перевод байтов в мегабиты)
	freeBandwidthBytes := totalBandwidth - usedBandwidth
	if totalBandwidth > 0 {
		// Переводим байты в мегабиты: ×8 бит/байт, ÷(1024²) бит/Мбит
		freeBandwidthMbit := float64(freeBandwidthBytes) / (1024.0 * 1024.0)
		bandwidthUsagePercent := (usedBandwidth / totalBandwidth) * 100

		if bandwidthUsagePercent > networkUsageThreshold {
			fmt.Printf("Network bandwidth usage high: %.0f Mbit/s available\n", freeBandwidthMbit)
		}
	}
}
