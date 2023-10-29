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
	Successor  []TokenStateId
}

type ByteProductionRule struct {
	Weights           []ByteWeightedRule
	PreSampledWeights []uint8
	currentIndex      int
}

func (r *ProductionRule) EncodeTokens(tokenBytes map[Token]TokenStateId) ByteProductionRule {
	rule := ByteProductionRule{
		Weights: make([]ByteWeightedRule, len(r.Weights), len(r.Weights)),
	}

	total := 0.0
	for w := 0; w < len(r.Weights); w++ {
		wt := r.Weights[w]
		encodedTokens := make([]TokenStateId, len(wt.Tokens), len(wt.Tokens))
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

	rule.PreSample()
	return rule
}

func (bp *ByteProductionRule) RandomizeWeights(delta float64) {
	currentWeights := make([]float64, len(bp.Weights), len(bp.Weights))
	for i := 0; i < len(bp.Weights); i++ {
		currentWeights[i] = bp.Weights[i].UpperLimit - bp.Weights[i].LowerLimit
	}

	total := 0.0
	for i := 0; i < len(bp.Weights); i++ {
		random := rand.Float64()
		_, tokens := bp.findSuccessorByProbability(random)
		for _, token := range tokens {
			currentWeights[token] += delta - rand.Float64()*2*delta
		}

		bp.Weights[i].LowerLimit = total
		total += currentWeights[i]
		bp.Weights[i].UpperLimit = total
	}
	bp.PreSample()
}

func (bp *ByteProductionRule) PreSample() {
	if bp.PreSampledWeights == nil {
		bp.PreSampledWeights = make([]uint8, 256, 256)
	}
	for i := 0; i < 256; i++ {
		random := rand.Float64() * (bp.Weights[len(bp.Weights)-1].UpperLimit)
		index, _ := bp.findSuccessorByProbability(random)
		bp.PreSampledWeights[i] = index
	}
}

func (bp *ByteProductionRule) ChooseSuccessor() []TokenStateId {
	if bp.PreSampledWeights != nil {
		tokens := bp.Weights[bp.PreSampledWeights[bp.currentIndex]].Successor
		bp.currentIndex++
		if bp.currentIndex == len(bp.PreSampledWeights) {
			bp.currentIndex = 0
		}
		return tokens
	}
	random := rand.Float64() * (bp.Weights[len(bp.Weights)-1].UpperLimit)
	_, tokens := bp.findSuccessorByProbability(random)
	return tokens
}

func (bp *ByteProductionRule) findSuccessorByProbability(p float64) (uint8, []TokenStateId) {
	// Use binary search to find the successor
	lo, hi := 0, len(bp.Weights)
	for lo < hi {
		mid := (lo + hi) / 2
		if p < bp.Weights[mid].LowerLimit {
			hi = mid
		} else if p >= bp.Weights[mid].UpperLimit {
			lo = mid + 1
		} else {
			return uint8(mid), bp.Weights[mid].Successor
		}
	}
	return 0, []TokenStateId{}
}
