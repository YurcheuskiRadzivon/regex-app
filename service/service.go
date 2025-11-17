package service

import (
	"errors"
	"html"
	"sort"
	"strings"
)

type RegexService struct{}

func NewRegexService() *RegexService {
	return &RegexService{}
}

type ProcessResult struct {
	Pattern         string
	Text            string
	Highlighted     string
	ErrorMessage    string
	ParseOk         bool
	Matches         []Match
	TimedOut        bool
	ValidationError bool
}

var ErrEmptyText = errors.New("пустой текст для поиска")

func (s *RegexService) Process(pattern, text string) ProcessResult {
	res := ProcessResult{
		Pattern: pattern,
		Text:    text,
	}

	if strings.TrimSpace(pattern) == "" {
		res.ErrorMessage = "Регулярное выражение не должно быть пустым."
		res.ValidationError = true
		return res
	}
	if strings.TrimSpace(text) == "" {
		res.ErrorMessage = "Текст для поиска не должен быть пустым."
		res.ValidationError = true
		return res
	}

	if err := ValidatePattern(pattern); err != nil {
		res.ErrorMessage = "Выражение не является корректным регулярным: " + err.Error()
		res.ValidationError = true
		return res
	}

	ast, err := Parse(pattern)
	if err != nil {
		res.ErrorMessage = "Ошибка при разборе регулярного выражения: " + err.Error()
		res.ValidationError = true
		return res
	}
	res.ParseOk = true

	matchRes := RunMatch(ast, text)
	res.Matches = matchRes.Matches
	res.TimedOut = matchRes.TimedOut
	res.Highlighted = highlightMatches(text, matchRes.Matches)

	if len(res.Matches) == 0 {
		if res.ErrorMessage == "" {
			res.ErrorMessage = "Совпадений не найдено."
		}
	}

	return res
}

func highlightMatches(text string, matches []Match) string {
	if len(matches) == 0 {
		return html.EscapeString(text)
	}

	runes := []rune(text)

	type seg struct{ s, e int }
	segs := make([]seg, 0, len(matches))
	for _, m := range matches {
		s := m.Start
		e := m.End
		if s < 0 {
			s = 0
		}
		if e < s {
			e = s
		}
		if e > len(runes) {
			e = len(runes)
		}
		if s == e {
			continue
		}
		segs = append(segs, seg{s, e})
	}
	if len(segs) == 0 {
		return html.EscapeString(text)
	}

	sort.Slice(segs, func(i, j int) bool {
		if segs[i].s == segs[j].s {
			return segs[i].e < segs[j].e
		}
		return segs[i].s < segs[j].s
	})

	merged := []seg{segs[0]}
	for _, seg := range segs[1:] {
		last := &merged[len(merged)-1]
		if seg.s <= last.e {
			if seg.e > last.e {
				last.e = seg.e
			}
		} else {
			merged = append(merged, seg)
		}
	}

	var b strings.Builder
	cur := 0
	for _, m := range merged {
		if m.s > cur {
			b.WriteString(html.EscapeString(string(runes[cur:m.s])))
		}
		b.WriteString(`<mark>`)
		b.WriteString(html.EscapeString(string(runes[m.s:m.e])))
		b.WriteString(`</mark>`)
		cur = m.e
	}
	if cur < len(runes) {
		b.WriteString(html.EscapeString(string(runes[cur:])))
	}

	return b.String()
}
