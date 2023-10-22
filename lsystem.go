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

func (r *ProductionRule) String() string {
	var sb strings.Builder
	sb.WriteRune('"')
	sb.WriteString(string(r.Predecessor))
	sb.WriteRune('"')
	sb.WriteString(": `")
	for i, wt := range r.WeightedTokens {
		sb.WriteString(strconv.FormatFloat(wt.Probability, 'f', 2, 64))
		sb.WriteString(" ")
		for _, t := range wt.Tokens {
			sb.WriteString(string(t))
			sb.WriteString(" ")
		}
		if i != len(r.WeightedTokens)-1 {
			sb.WriteString("; ")
		}
	}
	sb.WriteString("`")
	return sb.String()
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
	Axiom     Token
	Rules     map[Token]*ProductionRule
	Variables TokenSet
	Constants TokenSet
	State     []Token
}

func NewLSystem(axiom Token, rulesMap map[Token]*ProductionRule, vars TokenSet, consts TokenSet) *LSystem {
	return &LSystem{
		Axiom:     axiom,
		Rules:     rulesMap,
		Variables: vars,
		Constants: consts,
		State:     []Token{axiom},
	}
}

func (l *LSystem) IsVariable(t Token) bool {
	return l.Variables.Contains(t)
}

func (l *LSystem) IsConstant(t Token) bool {
	return l.Constants.Contains(t)
}

func (l *LSystem) IsStatefulVariable(t Token) bool {
	firstRune := rune(string(t)[0])
	lastRune := rune(string(t)[len(t)-1])
	return lastRune >= '0' && lastRune <= '9' && firstRune >= 'A' && firstRune <= 'Z'
}

func (l *LSystem) parseStatefulVariable(t Token) (Token, int) {
	digits := make([]rune, 0, 2)
	var sb strings.Builder
	for _, r := range t {
		if r >= '0' && r <= '9' {
			digits = append(digits, r)
			continue
		}
		if len(digits) == 0 {
			sb.WriteRune(r)
			continue
		}
	}

	num := 0
	for _, d := range digits {
		num = num*10 + int(d-'0')
	}

	return Token(sb.String()), num
}

func (l *LSystem) applyRules(input []Token) []Token {
	output := make([]Token, 0, len(input)*2)
	for _, token := range input {

		if l.IsStatefulVariable(token) {
			variable, num := l.parseStatefulVariable(token)
			if num == 0 {
				num++
			}
			token = Token(string(variable) + strconv.Itoa(num-1))
		}

		rule, exists := l.Rules[token]
		if exists && l.IsVariable(token) {
			output = append(output, rule.ChooseSuccessor()...)
		} else {
			output = append(output, token)
		}
	}
	return output
}

func (l *LSystem) IterateUntil(n int) []Token {
	result := []Token{l.Axiom}
	for i := 0; i < n; i++ {
		result = l.applyRules(result)
	}
	return result
}

func (l *LSystem) Iterate(n int) []Token {
	result := l.State
	for i := 0; i < n; i++ {
		result = l.applyRules(result)
	}

	l.State = result
	return result
}

func (l *LSystem) IterateOnce() []Token {
	return l.Iterate(1)
}

func (l *LSystem) String() string {
	var sb strings.Builder
	for _, rule := range l.Rules {
		sb.WriteString(rule.String())
		sb.WriteString(",\n")
	}
	return sb.String()
}

func (l *LSystem) Reset() {
	l.State = []Token{l.Axiom}
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

func isVariable(t Token) bool {
	endsWithUnderscore := string(t)[len(t)-1] == '_'
	return isCapitalized(t) && !endsWithUnderscore
}

func ParseRules(rulesMap map[Token]string) (TokenSet, TokenSet, map[Token]*ProductionRule) {
	vars := make(TokenSet)
	consts := make(TokenSet)
	parsedRules := make(map[Token]*ProductionRule)

	for key, value := range rulesMap {
		if isVariable(key) {
			vars.Add(key)
		} else {
			consts.Add(key)
		}
		parsedRules[key] = NewProductionRule(key, ParseRule(value))

		for _, wt := range parsedRules[key].WeightedTokens {
			for _, token := range wt.Tokens {
				if isVariable(token) {
					// remove numbers from variables
					token = Token(strings.TrimRight(string(token), "0123456789"))
					vars.Add(token)
				} else {
					consts.Add(token)
				}
			}
		}
	}

	return vars, consts, parsedRules
}

func ParseState(state string) []Token {
	return symbolsToTokens(strings.Fields(state))
}
