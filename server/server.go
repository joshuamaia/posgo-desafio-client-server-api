package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	apiURL       = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
	dbPath       = "cotacao.db"
	timeoutAPI   = 200 * time.Millisecond
	timeoutDB    = 10 * time.Millisecond
	clientTimeout = 300 * time.Millisecond
)

type Cotacao struct {
	Usdbrl struct {
		Ask        string `json:"ask"`
		Bid        string `json:"bid"`
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		CreateDate string `json:"create_date"`
		High       string `json:"high"`
		Low        string `json:"low"`
		Name       string `json:"name"`
		PctChange  string `json:"pctChange"`
		Timestamp  string `json:"timestamp"`
		VarBid     string `json:"varBid"`
	} `json:"USDBRL"`
}

func fetchCotacaoAtual(ctx context.Context) (Cotacao, error) {
	client := http.Client{Timeout: timeoutAPI}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return Cotacao{}, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return Cotacao{}, err
	}
	defer resp.Body.Close()
	
	var cotacao Cotacao
	if err := json.NewDecoder(resp.Body).Decode(&cotacao); err != nil {
		return Cotacao{}, err
	}

	return cotacao, nil
}

func saveDatabase(ctx context.Context, db *sql.DB, cotacao Cotacao) error {
	fmt.Println(cotacao)
	_, err := db.ExecContext(ctx, "INSERT INTO cotacoes (code, codein, name, high, low, varBid, pctChange, bid, ask, timestamp, create_date) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", 
	cotacao.Usdbrl.Code, cotacao.Usdbrl.Codein, cotacao.Usdbrl.Name, cotacao.Usdbrl.High, cotacao.Usdbrl.Low,
	cotacao.Usdbrl.VarBid, cotacao.Usdbrl.PctChange, cotacao.Usdbrl.Bid, cotacao.Usdbrl.Ask,
	cotacao.Usdbrl.Timestamp, cotacao.Usdbrl.CreateDate)
	return err
}

func main() {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS cotacoes (id INTEGER PRIMARY KEY, valor REAL)")
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), clientTimeout)
		defer cancel()

		cotacao, err := fetchCotacaoAtual(ctx)
		if err != nil {
			http.Error(w, "Erro ao obter cotação", http.StatusInternalServerError)
			return
		}

		ctx, cancel = context.WithTimeout(r.Context(), timeoutDB)
		defer cancel()
		if err := saveDatabase(ctx, db, cotacao); err != nil {
			log.Println("Erro ao salvar a cotação no banco:", err)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cotacao.Usdbrl.Bid)
	})

	log.Println("Servidor rodando na porta 8080...")
	http.ListenAndServe(":8080", nil)
}
