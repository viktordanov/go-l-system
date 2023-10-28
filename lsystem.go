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

	TokenBytes map[Token]BytePair
	BytesToken map[BytePair]Token
	ByteRules  [65535]ByteProductionRule
	BufferPool *BufferPool
	MemPool    *MemPool
}

func NewLSystem(axiom Token, rulesMap map[Token]ProductionRule, vars TokenSet, consts TokenSet) *LSystem {
	lSystem := &LSystem{
		Axiom:      axiom,
		Rules:      rulesMap,
		Variables:  vars,
		Constants:  consts,
		BufferPool: NewBufferPool(32),
		MemPool:    NewMemPool(32),
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
			// active encode base variable
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

func (l *LSystem) applyRules() {
	input := l.BufferPool.GetSwap()

	for _, token := range input.BytePairs[0:input.Len] {
		numPart := token.Second()
		if numPart != 0 {
			numPart--
		}

		newToken := NewBytePair(token.First(), numPart)
		rules := l.ByteRules[newToken]
		if rules.Weights == nil {
			l.BufferPool.Append(newToken)
			continue
		}

		l.BufferPool.AppendSlice(rules.ChooseSuccessor())
	}
	l.BufferPool.Swap()
	l.BufferPool.ResetWritingHead()
}

func (l *LSystem) IterateUntil(n int) []BytePair {
	l.Reset()
	for i := 0; i < n; i++ {
		l.applyRules()
	}
	return l.BufferPool.GetSwap().BytePairs[:l.BufferPool.GetSwap().Len]
}

func (l *LSystem) Iterate(n int) []BytePair {
	for i := 0; i < n; i++ {
		l.applyRules()
	}

	return l.BufferPool.GetSwap().BytePairs[:l.BufferPool.GetSwap().Len]
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
	l.BufferPool.Reset()
	l.BufferPool.GetSwap().BytePairs[0] = l.TokenBytes[l.Axiom]
	l.BufferPool.GetSwap().Len = 1
}
