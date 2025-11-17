package service

import "fmt"

type parser struct {
	runes []rune
	pos   int
}

func (p *parser) Parse(pattern string) (*Node, error) {
	p = &parser{runes: []rune(pattern)}
	node, err := p.parseAlternation()
	if err != nil {
		return nil, err
	}
	if !p.eof() {
		return nil, fmt.Errorf("лишние символы после позиции %d", p.pos)
	}
	return node, nil
}

func (p *parser) eof() bool {
	return p.pos >= len(p.runes)
}

func (p *parser) peek() rune {
	if p.eof() {
		return 0
	}
	return p.runes[p.pos]
}

func (p *parser) next() rune {
	ch := p.peek()
	p.pos++
	return ch
}

func (p *parser) parseAlternation() (*Node, error) {
	var branches []*Node

	left, err := p.parseConcatenation()
	if err != nil {
		return nil, err
	}
	branches = append(branches, left)

	for !p.eof() && p.peek() == '|' {
		p.next()
		right, err := p.parseConcatenation()
		if err != nil {
			return nil, err
		}
		branches = append(branches, right)
	}

	if len(branches) == 1 {
		return branches[0], nil
	}
	return &Node{
		Kind:     NodeAlternation,
		Children: branches,
	}, nil
}

func (p *parser) parseConcatenation() (*Node, error) {
	var nodes []*Node

	for !p.eof() {
		ch := p.peek()
		if ch == ')' || ch == '|' {
			break
		}
		n, err := p.parseQuantified()
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, n)
	}

	if len(nodes) == 0 {
		return &Node{Kind: NodeSequence, Children: nil}, nil
	}
	if len(nodes) == 1 {
		return nodes[0], nil
	}
	return &Node{
		Kind:     NodeSequence,
		Children: nodes,
	}, nil
}

func (p *parser) parseQuantified() (*Node, error) {
	atom, err := p.parseAtom()
	if err != nil {
		return nil, err
	}
	if p.eof() {
		return atom, nil
	}
	ch := p.peek()
	var q *Quantifier
	switch ch {
	case '*':
		p.next()
		q = &Quantifier{Min: 0, Max: -1, Greedy: true}
	case '+':
		p.next()
		q = &Quantifier{Min: 1, Max: -1, Greedy: true}
	case '?':
		p.next()
		q = &Quantifier{Min: 0, Max: 1, Greedy: true}
	default:
		return atom, nil
	}
	return &Node{
		Kind:  NodeQuantifier,
		Quant: q,
		Inner: atom,
	}, nil
}

func (p *parser) parseAtom() (*Node, error) {
	if p.eof() {
		return nil, fmt.Errorf("ожидался атом, но строка закончилась")
	}
	ch := p.next()

	switch ch {
	case '.':
		return &Node{Kind: NodeDot}, nil
	case '(':
		inner, err := p.parseAlternation()
		if err != nil {
			return nil, err
		}
		if p.eof() || p.peek() != ')' {
			return nil, fmt.Errorf("ожидалась ')'")
		}
		p.next()
		return &Node{Kind: NodeGroup, Children: []*Node{inner}}, nil
	case '[':
		return p.parseCharClass()
	case '\\':
		return p.parseEscape()
	default:
		return &Node{Kind: NodeLiteral, Literal: ch}, nil
	}
}

func (p *parser) parseEscape() (*Node, error) {
	if p.eof() {
		return nil, fmt.Errorf("экранирование в конце строки")
	}
	ch := p.next()
	switch ch {
	case 'd', 'w', 's':
		return &Node{Kind: NodeSpecialClass, Special: ch}, nil
	default:
		return &Node{Kind: NodeLiteral, Literal: ch}, nil
	}
}

func (p *parser) parseCharClass() (*Node, error) {
	cc := &CharClass{}
	if !p.eof() && p.peek() == '^' {
		cc.Negate = true
		p.next()
	}

	var last rune
	hasLast := false

	for !p.eof() && p.peek() != ']' {
		ch := p.next()
		if ch == '\\' {
			if p.eof() {
				return nil, fmt.Errorf("неполное экранирование в []")
			}
			ch = p.next()
			cc.Literals = append(cc.Literals, ch)
			last = ch
			hasLast = true
			continue
		}
		if ch == '-' && hasLast && p.peek() != ']' {
			if p.eof() {
				return nil, fmt.Errorf("незавершённый диапазон в []")
			}
			next := p.next()
			cc.Ranges = append(cc.Ranges, [2]rune{last, next})
			hasLast = false
			continue
		}
		cc.Literals = append(cc.Literals, ch)
		last = ch
		hasLast = true
	}

	if p.eof() || p.peek() != ']' {
		return nil, fmt.Errorf("незакрытый класс символов []")
	}
	p.next()

	return &Node{
		Kind:  NodeCharClass,
		Class: cc,
	}, nil
}
