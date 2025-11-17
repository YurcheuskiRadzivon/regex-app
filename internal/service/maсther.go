package service

const maxSteps = 20000

type matcher struct {
	steps int
}

func (m *matcher) RunMatch(root *Node, text string) MatchResult {
	m = &matcher{}
	runes := []rune(text)
	var matches []Match

	for start := 0; start <= len(runes); start++ {
		if m.steps > maxSteps {
			return MatchResult{Matches: matches, TimedOut: true}
		}
		ends := m.matchNode(root, runes, start)
		for _, end := range ends {
			if end > start {
				matches = append(matches, Match{Start: start, End: end})
			}
		}
	}

	return MatchResult{Matches: matches, TimedOut: m.steps > maxSteps}
}

func (m *matcher) matchNode(node *Node, runes []rune, pos int) []int {
	if m.steps > maxSteps {
		return nil
	}
	m.steps++

	switch node.Kind {
	case NodeLiteral:
		if pos < len(runes) && runes[pos] == node.Literal {
			return []int{pos + 1}
		}
		return nil
	case NodeDot:
		if pos < len(runes) {
			return []int{pos + 1}
		}
		return nil
	case NodeSpecialClass:
		if pos < len(runes) && matchSpecial(node.Special, runes[pos]) {
			return []int{pos + 1}
		}
		return nil
	case NodeCharClass:
		if pos < len(runes) && matchCharClass(node.Class, runes[pos]) {
			return []int{pos + 1}
		}
		return nil
	case NodeGroup:
		if len(node.Children) == 0 {
			return []int{pos}
		}
		return m.matchNode(node.Children[0], runes, pos)
	case NodeSequence:
		return m.matchSequence(node.Children, runes, pos)
	case NodeAlternation:
		var result []int
		for _, c := range node.Children {
			ends := m.matchNode(c, runes, pos)
			result = append(result, ends...)
		}
		return uniqueInts(result)
	case NodeQuantifier:
		return m.matchQuantifier(node, runes, pos)
	default:
		return nil
	}
}

func (m *matcher) matchSequence(children []*Node, runes []rune, pos int) []int {
	positions := []int{pos}
	for _, child := range children {
		var nextPositions []int
		for _, p := range positions {
			ends := m.matchNode(child, runes, p)
			nextPositions = append(nextPositions, ends...)
		}
		if len(nextPositions) == 0 {
			return nil
		}
		positions = uniqueInts(nextPositions)
	}
	return positions
}

func (m *matcher) matchQuantifier(node *Node, runes []rune, pos int) []int {
	q := node.Quant
	var results []int

	current := []int{pos}
	results = append(results, pos)

	for reps := 1; q.Max < 0 || reps <= q.Max; reps++ {
		var next []int
		for _, p := range current {
			ends := m.matchNode(node.Inner, runes, p)
			next = append(next, ends...)
		}
		if len(next) == 0 {
			break
		}
		current = uniqueInts(next)
		for _, p := range current {
			results = append(results, p)
		}
	}

	valid := uniqueInts(results)
	var filtered []int
	for _, end := range valid {
		if countReps(node, runes, pos, end) >= q.Min {
			filtered = append(filtered, end)
		}
	}
	return filtered
}

func countReps(node *Node, runes []rune, start, end int) int {
	cnt := 0
	pos := start
	for pos < end {
		m := &matcher{}
		ends := m.matchNode(node.Inner, runes, pos)
		if len(ends) == 0 {
			break
		}

		minEnd := ends[0]

		for _, e := range ends {
			if e < minEnd {
				minEnd = e
			}
		}
		pos = minEnd
		cnt++
	}
	return cnt
}

func matchCharClass(cc *CharClass, r rune) bool {
	match := false
	for _, lit := range cc.Literals {
		if r == lit {
			match = true
			break
		}
	}
	if !match {
		for _, rg := range cc.Ranges {
			if r >= rg[0] && r <= rg[1] {
				match = true
				break
			}
		}
	}
	if cc.Negate {
		return !match
	}
	return match
}

func matchSpecial(kind rune, r rune) bool {
	switch kind {
	case 'd':
		return r >= '0' && r <= '9'
	case 'w':
		return (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '_'
	case 's':
		return r == ' ' || r == '\t' || r == '\n' || r == '\r'
	default:
		return false
	}
}

func uniqueInts(a []int) []int {
	m := make(map[int]struct{}, len(a))
	var res []int
	for _, v := range a {
		if _, ok := m[v]; !ok {
			m[v] = struct{}{}
			res = append(res, v)
		}
	}
	return res
}
