package lsystem

import (
	"pgregory.net/rand"
	"strconv"
	"strings"
)

type WeightedRule struct {
	Probability float64
	Tokens      []Token
}

type ProductionRule struct {
	Predecessor Token
	Weights     []WeightedRule
}

func (r *ProductionRule) String() string {
	var sb strings.Builder
	sb.WriteRune('"')
	sb.WriteString(string(r.Predecessor))
	sb.WriteRune('"')
	sb.WriteString(": `")
	for i, wt := range r.Weights {
		sb.WriteString(strconv.FormatFloat(wt.Probability, 'f', 2, 64))
		sb.WriteString(" ")
		for _, t := range wt.Tokens {
			sb.WriteString(string(t))
			sb.WriteString(" ")
		}
		if i != len(r.Weights)-1 {
			sb.WriteString("; ")
		}
	}
	sb.WriteString("`")
	return sb.String()
}

func NewProductionRule(predecessor Token, weights []WeightedRule) ProductionRule {
	return ProductionRule{
		Predecessor: predecessor,
		Weights:     weights,
	}
}

func (r *ProductionRule) ChooseSuccessor() []Token {
	total := 0.0
	for _, wt := range r.Weights {
		total += wt.Probability
	}
	random := rand.Float64() * total
	for _, wt := range r.Weights {
		random -= wt.Probability
		if random < 0 {
			return wt.Tokens
		}
	}
	return []Token{}
}

type ByteWeightedRule struct {
	LowerLimit float64
	UpperLimit float64
	Successor  []BytePair
}

type ByteProductionRule struct {
	Predecessor BytePair
	Weights     []ByteWeightedRule
}

func (r *ProductionRule) encodeTokens(tokenBytes map[Token]BytePair) ByteProductionRule {
	rule := ByteProductionRule{
		Predecessor: tokenBytes[r.Predecessor],
		Weights:     make([]ByteWeightedRule, len(r.Weights), len(r.Weights)),
	}

	total := 0.0
	for w := 0; w < len(r.Weights); w++ {
		wt := r.Weights[w]
		encodedTokens := make([]BytePair, len(wt.Tokens), len(wt.Tokens))
		for i := len(wt.Tokens) - 1; i >= 0; i-- {
			t := wt.Tokens[i]
			encodedTokens[i] = tokenBytes[t]
		}
		rule.Weights[w] = ByteWeightedRule{
			Successor: encodedTokens,
		}
		rule.Weights[w].LowerLimit = total
		total += wt.Probability
		rule.Weights[w].UpperLimit = total
	}

	return rule
}

func (bp *ByteProductionRule) ChooseSuccessor() []BytePair {
	random := rand.Float64()
	return bp.findSuccessorByProbability(random)
}

func (bp *ByteProductionRule) findSuccessorByProbability(p float64) []BytePair {
	// Use binary search to find the successor
	lo, hi := 0, len(bp.Weights)
	for lo < hi {
		mid := (lo + hi) / 2
		if p < bp.Weights[mid].LowerLimit {
			hi = mid
		} else if p >= bp.Weights[mid].UpperLimit {
			lo = mid + 1
		} else {
			return bp.Weights[mid].Successor
		}
	}
	return []BytePair{}
}
