package lsystem

import (
	"strconv"
	"strings"
)

func ParseRule(str string) []WeightedRule {
	groups := strings.Split(strings.ReplaceAll(str, "\n", ""), ";")
	var weightedTokens []WeightedRule

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

func ParseRules(rulesMap map[Token]string) (TokenSet, TokenSet, map[Token]ProductionRule) {
	vars := make(TokenSet)
	consts := make(TokenSet)
	parsedRules := make(map[Token]ProductionRule)

	for key, value := range rulesMap {
		if isVariable(key) {
			vars.Add(key)
		} else {
			consts.Add(key)
		}
		parsedRules[key] = NewProductionRule(key, ParseRule(value))

		for _, wt := range parsedRules[key].Weights {
			for _, token := range wt.Tokens {
				if isVariable(token) {
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

func tryParseStatefulVariable(t Token) (variable string, num uint8, ok bool) {
	var sb strings.Builder
	cumulativeNumber := 0

	if t[len(t)-1] < '0' || t[len(t)-1] > '9' {
		return "", 0, false
	}

	for _, r := range t {
		if r >= '0' && r <= '9' {
			cumulativeNumber = cumulativeNumber*10 + int(r-'0')
			continue
		}
		if cumulativeNumber == 0 {
			sb.WriteRune(r)
			continue
		}
	}
	if cumulativeNumber == 0 {
		return "", 0, false
	}
	if cumulativeNumber > 255 {
		cumulativeNumber = 255
	}
	num = uint8(cumulativeNumber)

	return sb.String(), num, true
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
