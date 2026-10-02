package z01

import (
	"bytes"
	"html/template"
	"math/rand"
	"net/http"
	"os"
	"strings"
)

type PageData struct {
	Text   string
	Banner string
	Splash string
	Result string
}

func randomSplash() string {
	data, err := os.ReadFile("static/splash.txt")
	if err != nil {
		return "A NEW WORLD AWAITS"
	}

	var lines []string
	for _, line := range strings.Split(string(data), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			lines = append(lines, line)
		}
	}
	if len(lines) == 0 {
		return "A NEW WORLD AWAITS"
	}
	return lines[rand.Intn(len(lines))]
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		errorPage(w, http.StatusNotFound, "Page not found")
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		errorPage(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	renderPage(w, http.StatusOK, PageData{Banner: DefaultBanner, Splash: randomSplash()})
}

func AsciiArtHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/ascii-art" {
		errorPage(w, http.StatusNotFound, "Page not found")
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		errorPage(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, MaxInputSize)
	if err := r.ParseForm(); err != nil {
		errorPage(w, http.StatusBadRequest, "Invalid form data")
		return
	}

	text := normalizeText(r.FormValue("text"))
	banner := r.FormValue("banner")
	if banner == "" {
		banner = DefaultBanner
	}
	if !ValidInput(text, banner) {
		errorPage(w, http.StatusBadRequest, "Invalid input")
		return
	}

	result, err := GenerateASCII(text, banner)
	if err != nil {
		if os.IsNotExist(err) {
			errorPage(w, http.StatusNotFound, "Missing Banner file")
		} else {
			errorPage(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	renderPage(w, http.StatusOK, PageData{Text: text, Banner: banner, Splash: randomSplash(), Result: result})
}

func renderPage(w http.ResponseWriter, status int, data PageData) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		if os.IsNotExist(err) {
			errorPage(w, http.StatusNotFound, "Missing page template file")
		} else {
			errorPage(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		errorPage(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(body.Bytes())
}

func errorPage(w http.ResponseWriter, status int, message string) {
	tmpl, err := template.ParseFiles("templates/error.html")
	if err != nil {
		http.Error(w, message, status)
		return
	}

	data := struct {
		StatusCode int
		Message    string
	}{StatusCode: status, Message: message}
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		http.Error(w, message, status)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(body.Bytes())
}
