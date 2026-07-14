package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type payRequest struct {
	OrderNo    string `json:"orderNo"`
	AmountCent int64  `json:"amountCent"`
}

type payResponse struct {
	TradeNo string `json:"tradeNo"`
	Status  string `json:"status"`
}

type refundRequest struct {
	TradeNo    string `json:"tradeNo"`
	OrderNo    string `json:"orderNo"`
	AmountCent int64  `json:"amountCent"`
}

type refundResponse struct {
	TradeNo string `json:"tradeNo"`
	Status  string `json:"status"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/pay", handlePay)
	mux.HandleFunc("/refund", handleRefund)
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	server := &http.Server{
		Addr:              ":9001",
		Handler:           mux,
		ReadHeaderTimeout: 3 * time.Second,
	}
	log.Println("mock payment listening on http://127.0.0.1:9001")
	log.Fatal(server.ListenAndServe())
}

func handlePay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"message": "只允许 POST"})
		return
	}

	var req payRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "JSON 格式错误"})
		return
	}
	if req.OrderNo == "TIMEOUT" {
		time.Sleep(3 * time.Second)
	}
	if req.OrderNo == "FAIL" {
		writeJSON(w, http.StatusBadGateway, map[string]string{"message": "模拟支付渠道故障"})
		return
	}
	if req.OrderNo == "" || req.AmountCent <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "订单号和金额不合法"})
		return
	}

	writeJSON(w, http.StatusOK, payResponse{
		TradeNo: fmt.Sprintf("MOCK-%s", req.OrderNo),
		Status:  "paid",
	})
}

func handleRefund(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"message": "只允许 POST"})
		return
	}

	var req refundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "JSON 格式错误"})
		return
	}
	if req.OrderNo == "TIMEOUT" {
		time.Sleep(3 * time.Second)
	}
	if req.OrderNo == "FAIL" {
		writeJSON(w, http.StatusBadGateway, map[string]string{"message": "模拟退款渠道故障"})
		return
	}
	if req.OrderNo == "" || req.TradeNo == "" || req.AmountCent <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "订单号、交易号和金额不合法"})
		return
	}

	writeJSON(w, http.StatusOK, refundResponse{
		TradeNo: req.TradeNo,
		Status:  "refunded",
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("write response: %v", err)
	}
}
