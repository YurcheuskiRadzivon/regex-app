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
	NodeSpecialClass
)

type CharClass struct {
	Negate   bool
	Ranges   [][2]rune
	Literals []rune
}

type Quantifier struct {
	Min    int
	Max    int
	Greedy bool
}

type Node struct {
	Kind NodeKind

	Literal rune

	Special rune

	Class *CharClass

	Children []*Node

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
