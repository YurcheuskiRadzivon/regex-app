package service

type NodeKind int

const (
	NodeLiteral NodeKind = iota
	NodeDot
	NodeCharClass
	NodeSequence
	NodeAlternation
	NodeGroup
	NodeQuantifier
	NodeSpecialClass // \d, \w, \s
)

type CharClass struct {
	Negate   bool
	Ranges   [][2]rune // от–до
	Literals []rune
}

type Quantifier struct {
	Min    int
	Max    int // -1 = бесконечность
	Greedy bool
}

type Node struct {
	Kind NodeKind

	// Literal
	Literal rune

	// SpecialClass: 'd','w','s'
	Special rune

	// CharClass
	Class *CharClass

	// Sequence / Alternation / Group
	Children []*Node

	// Quantifier
	Quant *Quantifier
	Inner *Node
}

type Match struct {
	Start int
	End   int
}

type MatchResult struct {
	Matches  []Match
	TimedOut bool
}
