package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	conf "github.com/gate_entry/config"
)

var logger *log.Logger
var config = conf.Get()

func initLogger() {
	logFile := filepath.Join(config.Log.Path, "proxy.log")
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Cannot write to log file: %v", err)
	}
	logger = log.New(io.MultiWriter(os.Stdout, f), "", log.LstdFlags)
	logger.Printf("Log initialized at %s [%s]", logFile, config.Log.Level)
}

func extractUriParams(q url.Values) map[string]string {
	var res = make(map[string]string)
	if len(q) > 0 {
		for k, v := range q {
			res[k] = strings.Join(v, ",")
		}
	}
	return res
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/favicon.ico" {
		http.NotFound(w, r)
		return
	}
	query := r.URL.RawQuery

	logger.Printf("[REQ] Method: %s | URI: %s | [QUERY] %s", r.Method, r.RequestURI, query)

	var bodyCopy []byte
	if r.Body != nil {
		bodyCopy, _ = io.ReadAll(r.Body)
		r.Body = io.NopCloser(strings.NewReader(string(bodyCopy))) // 复用 body
	}
	if len(bodyCopy) > 0 {
		logger.Printf("[REQ] [BODY] %s", string(bodyCopy))
	}

	targetURL := strings.TrimRight(config.Proxy.TargetURI, "/") + r.URL.Path
	if query != "" {
		targetURL += "?" + query
	}

	paramMap := extractUriParams(r.URL.Query())
	cmd := paramMap["a"]
	scanType := paramMap["scanType"]
	scanContent := paramMap["scanContent"]

	if cmd == "checkticket" && scanType == "2" {
		if strings.HasPrefix(scanContent, "so-") {
			targetURL = "https://qrcinema.cn/general_api/api/gate/check?modify_status=modify&qk=" + scanContent
		}
	}

	req, err := http.NewRequest(r.Method, targetURL, strings.NewReader(string(bodyCopy)))
	if err != nil {
		http.Error(w, "Failed to create request: "+err.Error(), http.StatusInternalServerError)
		logger.Println("Error creating request:", err)
		return
	}
	req.Header = r.Header.Clone()

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Forwarding failed: "+err.Error(), http.StatusBadGateway)
		logger.Println("Forwarding error:", err)
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response", http.StatusInternalServerError)
		logger.Println("Error reading response:", err)
		return
	}

	respBodyStr := string(respBody)
	if strings.Contains(respBodyStr, "<html") || strings.Contains(respBodyStr, "<!DOCTYPE html") || strings.Contains(respBodyStr, "<HTML") {
		logger.Printf("[RESP_BODY] [SKIPPED: HTML response]")
	} else {
		logger.Printf("[RESP_BODY] %s", respBodyStr)
	}

	for k, v := range resp.Header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}
	w.WriteHeader(resp.StatusCode)
	w.Write(respBody)

	logger.Printf("[RESP] → %s (%d)", targetURL, resp.StatusCode)
}

func main() {
	initLogger()

	port := config.Proxy.ServerPort
	logger.Printf("Proxy server started on :%d", port)
	logger.Printf("Forwarding all requests to %s", config.Proxy.TargetURI)
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))
}
