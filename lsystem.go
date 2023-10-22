package lsystem

import (
	"math/rand"
	"strconv"
	"strings"
)

type Token string

type TokenSet map[Token]struct{}

func (ts TokenSet) Contains(t Token) bool {
	_, exists := ts[t]
	return exists
}

func (ts TokenSet) Add(t Token) {
	ts[t] = struct{}{}
}

func (ts TokenSet) AsSlice() []Token {
	slice := make([]Token, 0, len(ts))
	for t := range ts {
		slice = append(slice, t)
	}
	return slice
}

type ProductionRule struct {
	Predecessor    Token
	WeightedTokens []struct {
		Probability float64
		Tokens      []Token
	}
}

func NewProductionRule(predecessor Token, weightedTokens []struct {
	Probability float64
	Tokens      []Token
}) *ProductionRule {
	return &ProductionRule{
		Predecessor:    predecessor,
		WeightedTokens: weightedTokens,
	}
}

func (r *ProductionRule) ChooseSuccessor() []Token {
	total := 0.0
	for _, wt := range r.WeightedTokens {
		total += wt.Probability
	}
	random := rand.Float64() * total
	for _, wt := range r.WeightedTokens {
		random -= wt.Probability
		if random < 0 {
			return wt.Tokens
		}
	}
	return []Token{}
}

type LSystem struct {
	Axiom     []Token
	Rules     map[Token]*ProductionRule
	Variables TokenSet
	Constants TokenSet
}

func NewLSystem(axiom Token, rulesMap map[Token]*ProductionRule, vars TokenSet, consts TokenSet) *LSystem {
	return &LSystem{
		Axiom:     []Token{axiom},
		Rules:     rulesMap,
		Variables: vars,
		Constants: consts,
	}
}

func (l *LSystem) IsVariable(t Token) bool {
	return l.Variables.Contains(t)
}

func (l *LSystem) IsConstant(t Token) bool {
	return l.Constants.Contains(t)
}

// Updated applyRules method to replace only variables
func (l *LSystem) applyRules(input []Token) []Token {
	output := make([]Token, 0)
	for _, token := range input {
		rule, exists := l.Rules[token]
		if exists && l.IsVariable(token) {
			output = append(output, rule.ChooseSuccessor()...)
		} else {
			output = append(output, token)
		}
	}
	return output
}

func (l *LSystem) Iterate(n int) []Token {
	result := l.Axiom
	for i := 0; i < n; i++ {
		result = l.applyRules(result)
	}
	return result
}

func ParseRule(str string) []struct {
	Probability float64
	Tokens      []Token
} {
	groups := strings.Split(strings.ReplaceAll(str, "\n", ""), ";")
	var weightedTokens []struct {
		Probability float64
		Tokens      []Token
	}

	for _, group := range groups {
		if strings.TrimSpace(group) == "" {
			continue
		}
		tokens := strings.Fields(group)
		weight, err := strconv.ParseFloat(tokens[0], 64)
		if err != nil {
			continue
		}
		weightedTokens = append(weightedTokens, struct {
			Probability float64
			Tokens      []Token
		}{
			Probability: weight,
			Tokens:      symbolsToTokens(tokens[1:]),
		})
	}
	return weightedTokens
}

func symbolsToTokens(symbols []string) []Token {
	var tokens []Token
	for _, symbol := range symbols {
		tokens = append(tokens, Token(symbol))
	}
	return tokens
}

func isCapitalized(t Token) bool {
	firstLetter := string(t)[0]
	return firstLetter >= 'A' && firstLetter <= 'Z'
}

func ParseRules(rulesMap map[Token]string) (TokenSet, TokenSet, map[Token]*ProductionRule) {
	vars := make(TokenSet)
	consts := make(TokenSet)
	parsedRules := make(map[Token]*ProductionRule)

	for key, value := range rulesMap {
		if isCapitalized(key) {
			vars.Add(key)
		} else {
			consts.Add(key)
		}
		parsedRules[key] = NewProductionRule(key, ParseRule(value))

		for _, wt := range parsedRules[key].WeightedTokens {
			for _, token := range wt.Tokens {
				if isCapitalized(token) {
					vars.Add(token)
				} else {
					consts.Add(token)
				}
			}
		}
	}

	return vars, consts, parsedRules
}
