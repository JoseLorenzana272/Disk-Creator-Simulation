package main

import (
	"archivos_pro1/Analyzer"
	"encoding/json"
	"net/http"
)

type CodeRequest struct {
	Code string `json:"code"`
}

type CodeResponse struct {
	Outputs interface{} `json:"output"`
}

func runCodeHandler(w http.ResponseWriter, r *http.Request) {
	// Habilitar CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req CodeRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	outputs, err := Analyzer.Analyzer(req.Code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := CodeResponse{
		Outputs: outputs,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	http.HandleFunc("/run-code", runCodeHandler)
	http.ListenAndServe(":8080", nil)
}
