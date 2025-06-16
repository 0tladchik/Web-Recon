// Пакет pkg содержит основные модули для рекона.
package pkg

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// ScanResult представляет результат сканирования одного порта.
type ScanResult struct {
	Port   int
	IsOpen bool
}

// PortScanner запускает сканирование портов для указанного хоста.
// Использует горутины для параллельного сканирования, что значительно ускоряет процесс.
func PortScanner(host string, startPort, endPort int) []ScanResult {
	var results []ScanResult
	var wg sync.WaitGroup      // WaitGroup для ожидания завершения всех горутин.
	var mu sync.Mutex          // Мьютекс для безопасной записи в слайс results.
	
	// Ограничиваем количество одновременных горутин для предотвращения перегрузки.
	semaphore := make(chan struct{}, 100)

	for port := startPort; port <= endPort; port++ {
		wg.Add(1)
		semaphore <- struct{}{} // Занимаем место в семафоре

		go func(p int) {
			defer wg.Done()
			defer func() { <-semaphore }() // Освобождаем место по завершении

			address := fmt.Sprintf("%s:%d", host, p)
			// Использование DialTimeout - хорошая практика, чтобы не зависать надолго на "плохих" портах.
			conn, err := net.DialTimeout("tcp", address, 1*time.Second)

			if err == nil {
				conn.Close()
				mu.Lock() // Блокируем мьютекс перед доступом к общему ресурсу.
				results = append(results, ScanResult{Port: p, IsOpen: true})
				mu.Unlock() // Разблокируем мьютекс.
			}
		}(port)
	}

	wg.Wait() // Ожидаем завершения всех сканирований.
	return results
}