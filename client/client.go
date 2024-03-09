package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

const (
	serverURL      = "http://localhost:8080/cotacao" // Endereço do server.go
	clientTimeout  = 300 * time.Millisecond
	outputFilename = "cotacao.txt"
)

func fetchCotacao(ctx context.Context) (string, error) {
	client := http.Client{Timeout: clientTimeout}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, serverURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var bid string
	if err := json.NewDecoder(resp.Body).Decode(&bid); err != nil {
		return "", err
	}

	return bid, nil
}

func saveToFile(bid string) error {
	content := fmt.Sprintf("Dólar: %s", bid)
	return os.WriteFile(outputFilename, []byte(content), 0644)
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), clientTimeout)
	defer cancel()

	cotacao, err := fetchCotacao(ctx)
	if err != nil {
		fmt.Println("Erro ao obter cotação:", err)
		return
	}

	if err := saveToFile(cotacao); err != nil {
		fmt.Println("Erro ao salvar no arquivo:", err)
		return
	}

	fmt.Printf("Cotação atual salva em %s\n", outputFilename)
}
