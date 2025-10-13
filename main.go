package main

import (
	"bufio"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
)

var templates *template.Template

type PageData struct {
	ASCIIArt string
	Error    string
}

func main() {
	templates = template.Must(template.ParseGlob("templates/*"))

	// Serve static assets from /static/
	fs := http.FileServer(http.Dir("./templates"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Delegate to template rendering for root and GET
		pageData := PageData{}
		if r.Method == http.MethodGet && r.URL.Path == "/" {
			templates.ExecuteTemplate(w, "index.html", pageData)
			return
		}
		// Not found
		pageData.Error = "ERROR-404\nPage not found"
		templates.ExecuteTemplate(w, "index.html", pageData)
	})

	http.HandleFunc("/asciiart", posthandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	log.Printf("starting server on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func posthandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			http.Error(w, "ERROR-500\nInternal Server Error", http.StatusInternalServerError)
		}
	}()

	pageData := PageData{}

	if r.Method == "GET" {
		if r.URL.Path != "/" {
			pageData.Error = "ERROR-404\nPage not found("
			templates.ExecuteTemplate(w, "index.html", pageData)
			return
		}
		templates.ExecuteTemplate(w, "index.html", pageData)
		return
	}

	if r.Method == "POST" {
		text1 := r.FormValue("text")
		font := r.FormValue("fonts")
		text := ""
		if strings.Contains(text1, "\r\n") {
			text = strings.ReplaceAll(text1, "\r\n", "\\n")
		} else {
			text = text1
		}
		for _, v := range text {
			if !(v >= 32 && v <= 126) {
				pageData.Error = "ERROR-400\nBad request!"
				templates.ExecuteTemplate(w, "index.html", pageData)
				return
			}
		}
		file, err := os.Open(font + ".txt")
		if err != nil {
			pageData.Error = "ERROR-400\nBad Request!! \nPlease make sure you select a font."
			templates.ExecuteTemplate(w, "index.html", pageData)
			return
		}
		defer file.Close()

		Scanner := bufio.NewScanner(file)
		var lines []string
		for Scanner.Scan() {
			lines = append(lines, Scanner.Text())
		}

		asciiChrs := make(map[int][]string)
		dec := 31

		for _, line := range lines {
			if line == "" {
				dec++
			} else {
				asciiChrs[dec] = append(asciiChrs[dec], line)
			}
		}

		var c = ""
		for i := 0; i < len(text); i++ {
			if text[i] == 92 && text[i+1] == 110 {
				c = PrintArt(text[:i], asciiChrs) + PrintArt(text[i+2:], asciiChrs)
			}
		}
		if !strings.Contains(text, "\\n") {
			c = PrintArt(text, asciiChrs)
		}

		pageData.ASCIIArt = c
		templates.ExecuteTemplate(w, "index.html", pageData)
	}
}

func PrintArt(n string, y map[int][]string) string {
	a := []string{}
	for j := 0; j < len(y[32]); j++ {
		for _, letter := range n {
			a = append(a, y[int(letter)][j])
		}
		a = append(a, "\n")
	}
	b := strings.Join(a, "")
	return b
}
