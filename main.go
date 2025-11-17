package main

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"regexp-helper/service"
)

type MatchView struct {
	Index int
	From  int
	To    int
	Text  string
}

type PageData struct {
	Pattern      string
	Text         string
	ResultHTML   template.HTML
	ErrorMessage string
	ParseOk      bool
	HasMatches   bool
	HasMultiple  bool
	TimedOut     bool

	MatchList []MatchView
}

func main() {
	tmplPath := filepath.Join("templates", "index.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		log.Fatalf("Ошибка загрузки шаблона: %v", err)
	}

	svc := service.NewRegexService()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			data := PageData{}
			if err := tmpl.Execute(w, data); err != nil {
				log.Println("template error:", err)
			}
		case http.MethodPost:
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}
			pattern := r.FormValue("pattern")
			text := r.FormValue("text")

			res := svc.Process(pattern, text)

			data := PageData{
				Pattern:      pattern,
				Text:         text,
				ErrorMessage: res.ErrorMessage,
				ParseOk:      res.ParseOk,
				TimedOut:     res.TimedOut,
			}

			if res.Highlighted != "" {
				data.ResultHTML = template.HTML(res.Highlighted)
			}

			if len(res.Matches) > 0 {
				data.HasMatches = true
			}
			if len(res.Matches) > 1 {
				data.HasMultiple = true
			}

			// формируем список всех совпадений
			runes := []rune(text)
			for i, m := range res.Matches {
				from := m.Start
				to := m.End
				if from < 0 {
					from = 0
				}
				if to > len(runes) {
					to = len(runes)
				}
				if to < from {
					to = from
				}
				mv := MatchView{
					Index: i + 1,
					From:  from,
					To:    to,
					Text:  string(runes[from:to]),
				}
				data.MatchList = append(data.MatchList, mv)
			}

			if err := tmpl.Execute(w, data); err != nil {
				log.Println("template error:", err)
			}
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Println("Сервер запущен на http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
