package server

import (
	"html/template"
	"regexp-helper/internal/service"

	"github.com/gofiber/fiber/v2"
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

func NewRouter(
	app *fiber.App,
	svc *service.RegexService,

) {
	app.Get("/", func(c *fiber.Ctx) error {
		data := PageData{}
		return c.Render("index", data)
	})

	app.Post("/", func(c *fiber.Ctx) error {
		data := PageData{}

		if err := c.BodyParser(&data); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Bad Request")
		}

		pattern := c.FormValue("pattern")
		text := c.FormValue("text")

		res := svc.Process(pattern, text)

		data = PageData{
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

		return c.Render("index", data)
	})
}
