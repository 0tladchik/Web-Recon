// Главный пакет программы "Зеркало".
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	// Корректный импорт нашего локального пакета.
	"github.com/0tladchik/Web-Recon/pkg"
)

func main() {
	// --- Определение и парсинг флагов командной строки ---
	target := flag.String("d", "", "Целевой домен для сканирования (обязательный).")
	ports := flag.String("p", "1-1024", "Диапазон портов для сканирования (например, '80,443' или '1-1024').")
	wordlist := flag.String("w", "wordlist.txt", "Путь к файлу со списком субдоменов.")
	flag.Parse()

	if *target == "" {
		fmt.Println("Ошибка: не указан целевой домен. Используйте флаг -d.")
		flag.PrintDefaults() // Выводим справку по флагам.
		os.Exit(1)
	}

	fmt.Printf("[+] Запускаем разведку для цели: %s\n\n", *target)

	// --- Шаг 1: Поиск субдоменов ---
	fmt.Println("[*] Начинаем поиск субдоменов...")
	foundSubdomains := pkg.SubdomainScanner(*target, *wordlist)
	if len(foundSubdomains) > 0 {
		fmt.Printf("[+] Найдено %d субдоменов:\n", len(foundSubdomains))
		for _, sub := range foundSubdomains {
			fmt.Printf("  - %s\n", sub)
		}
	} else {
		fmt.Println("[-] Субдомены не найдены.")
	}
	fmt.Println()


	// --- Шаг 2: Сканирование портов ---
	fmt.Printf("[*] Начинаем сканирование портов для основного домена (%s)...\n", *target)
	startPort, endPort := parsePortRange(*ports)
	openPorts := pkg.PortScanner(*target, startPort, endPort)

	if len(openPorts) > 0 {
		fmt.Printf("[+] Найдены открытые порты:\n")
		for _, result := range openPorts {
			fmt.Printf("  - TCP/%d\n", result.Port)
		}
	} else {
		fmt.Println("[-] Открытые порты не найдены в указанном диапазоне.")
	}

	fmt.Println("\n[+] Разведка завершена.")
}

// Вспомогательная функция для парсинга диапазона портов.
func parsePortRange(rangeStr string) (int, int) {
	parts := strings.Split(rangeStr, "-")
	if len(parts) != 2 {
		// Упрощенная обработка ошибок для нашего случая.
		return 1, 1024
	}
	start, _ := strconv.Atoi(parts[0])
	end, _ := strconv.Atoi(parts[1])
	return start, end
}