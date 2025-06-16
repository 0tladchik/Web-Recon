package pkg

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"sync"
)

const (
	// Количество одновременных "воркеров" для разрешения DNS.
	// Подбирается экспериментально для оптимальной скорости.
	resolverWorkers = 50
)

// SubdomainScanner выполняет поиск субдоменов для целевого домена.
// Основной метод - брутфорс по словарю.
func SubdomainScanner(domain, wordlistPath string) []string {
	
	// Используем map для автоматической дедупликации найденных субдоменов.
	// struct{} не занимает памяти.
	foundSubdomains := make(map[string]struct{})
	var mu sync.Mutex

	// --- Конвейер для обработки ---
	jobs := make(chan string, resolverWorkers)
	results := make(chan string)
	var wg sync.WaitGroup

	// Запускаем воркеры, которые будут проверять резолвится ли домен.
	for i := 0; i < resolverWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for sub := range jobs {
				fqdn := fmt.Sprintf("%s.%s", sub, domain)
				ips, err := net.LookupHost(fqdn)
				if err == nil && len(ips) > 0 {
					results <- fqdn
				}
			}
		}()
	}

	// Горутина для сбора и сохранения результатов.
	go func() {
		for found := range results {
			mu.Lock()
			foundSubdomains[found] = struct{}{}
			mu.Unlock()
		}
	}()
	
	// TODO: Реализовать клиенты для внешних сервисов (VirusTotal, SecurityTrails). Пока что заглушка.
	// subdomainsFromAPIs := queryExternalAPIs(domain)
	
	// Читаем вордлист и отправляем задания в канал jobs.
	file, err := os.Open(wordlistPath)
	if err != nil {
		// В реальном инструменте здесь была бы более грамотная обработка ошибок.
		// Для демо - достаточно просто вывести в консоль.
		fmt.Printf("Ошибка открытия файла со словарем: %v\n", err)
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		jobs <- scanner.Text()
	}

	close(jobs)
	wg.Wait()
	close(results)

	// Преобразуем map в слайс для возврата.
	finalList := make([]string, 0, len(foundSubdomains))
	for sub := range foundSubdomains {
		finalList = append(finalList, sub)
	}

	return finalList
}