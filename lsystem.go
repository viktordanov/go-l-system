package lsystem

import (
	"strconv"
	"strings"
)

type LSystem struct {
	Axiom     Token
	Rules     map[Token]ProductionRule
	Variables TokenSet
	Constants TokenSet
	State     []BytePair

	TokenBytes map[Token]BytePair
	BytesToken map[BytePair]Token
	ByteRules  [65535]ByteProductionRule
}

func NewLSystem(axiom Token, rulesMap map[Token]ProductionRule, vars TokenSet, consts TokenSet) *LSystem {
	lSystem := &LSystem{
		Axiom:     axiom,
		Rules:     rulesMap,
		Variables: vars,
		Constants: consts,
		State:     []BytePair{},
	}

	lSystem.encodeTokens()
	lSystem.Reset()
	return lSystem
}

func (l *LSystem) encodeTokens() {
	l.TokenBytes = make(map[Token]BytePair)
	l.BytesToken = make(map[BytePair]Token)
	i := uint8(0)
	for t := range l.Variables {
		baseVar, numberState, isStateful := tryParseStatefulVariable(t)
		if isStateful {
			// first encode base variable
			baseVar := Token(baseVar)
			baseBytes, exists := l.TokenBytes[baseVar]
			if !exists {
				bytePair := NewBytePair(i, 0)
				l.TokenBytes[baseVar] = bytePair
				l.BytesToken[bytePair] = baseVar

				baseBytes = bytePair
			}

			bytePair := NewBytePair(baseBytes.First(), numberState)
			l.TokenBytes[Token(string(baseVar)+strconv.Itoa(int(numberState)))] = bytePair
			l.BytesToken[bytePair] = Token(string(baseVar) + strconv.Itoa(int(numberState)))

		} else {
			bytePair := NewBytePair(i, 0)
			l.TokenBytes[t] = bytePair
			l.BytesToken[bytePair] = t
		}
		i++
	}
	for t := range l.Constants {
		bytePair := NewBytePair(i, 0)
		l.TokenBytes[t] = bytePair
		l.BytesToken[bytePair] = t
		i++
	}

	l.ByteRules = [65535]ByteProductionRule{}
	for t, rule := range l.Rules {
		l.ByteRules[l.TokenBytes[t]] = rule.encodeTokens(l.TokenBytes)
	}
}

func (l *LSystem) DecodeBytes(bp []BytePair) []Token {
	result := make([]Token, 0, len(bp))
	for _, bytePair := range bp {
		v, exists := l.BytesToken[bytePair]
		if !exists {
			base := bytePair.First()
			v = Token(string(l.BytesToken[NewBytePair(base, 0)]) + strconv.Itoa(int(bytePair.Second())))
		}
		result = append(result, v)
	}
	return result
}

func (l *LSystem) IsVariable(t Token) bool {
	return l.Variables.Contains(t)
}

func (l *LSystem) IsConstant(t Token) bool {
	return l.Constants.Contains(t)
}

// TODO: Optimizations to do:
// 1) use an array for the rules and represent variables and constatns as int8s (indices in the array)
// to completely circumvent the map lookups
// 2) store a counter for each variable to avoid parsing the variable name and number every time
// -1 for no number, any other number means it's stateful
// 3) think about a better way to do random selection
// 4) implement a memory pool which automatically grows when necessary
// it will start off as 2 big preallocated arrays, one for the current state and one for the next state
// which will be swapped after each iteration and the other one will be overwritten in-place (store their length as well)
// Also make apply rules work in-place instead of allocating a new array and returning it

func (l *LSystem) applyRules(input []BytePair) []BytePair {
	output := make([]BytePair, 0, len(input)*2)

	for _, token := range input {
		numPart := token.Second()
		if numPart != 0 {
			numPart--
		}

		newToken := NewBytePair(token.First(), numPart)
		rules := l.ByteRules[newToken]
		if rules.Weights == nil {
			output = append(output, newToken)
			continue
		}
		output = append(output, rules.ChooseSuccessor()...)
	}
	return output
}

func (l *LSystem) IterateUntil(n int) []BytePair {
	result := []BytePair{l.TokenBytes[l.Axiom]}
	for i := 0; i < n; i++ {
		result = l.applyRules(result)
	}
	return result
}

func (l *LSystem) Iterate(n int) []BytePair {
	result := l.State
	for i := 0; i < n; i++ {
		result = l.applyRules(result)
	}

	l.State = result
	return result
}

func (l *LSystem) IterateOnce() []BytePair {
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
	l.State = []BytePair{l.TokenBytes[l.Axiom]}
}
