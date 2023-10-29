package lsystem

import (
	"strconv"
	"strings"
	"sync"
)

type LSystem struct {
	Axiom     Token
	Rules     map[Token]ProductionRule
	Variables TokenSet
	Constants TokenSet

	TokenBytes map[Token]BytePair
	BytesToken map[BytePair]Token
	ByteRules  [65535]ByteProductionRule
	MemPool    *MemPool
}

func NewLSystem(axiom Token, rulesMap map[Token]ProductionRule, vars TokenSet, consts TokenSet) *LSystem {
	lSystem := &LSystem{
		Axiom:     axiom,
		Rules:     rulesMap,
		Variables: vars,
		Constants: consts,
		MemPool:   NewMemPool(32),
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

func (l *LSystem) applyRules(n int) {
	var wg sync.WaitGroup
	for i := 0; i < threadCount; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			for j := 0; j < n; j++ {
				l.applyRulesOnce(l.MemPool.GetReadBuffer(i), l.MemPool.GetWriteBuffer(i))
				l.MemPool.Swap(i)
			}
		}(i)
	}

	wg.Wait()
}

func (l *LSystem) applyRulesOnce(input, output *Buffer) {
	for _, token := range input.BytePairs[:input.Len] {
		numPart := token.Second()
		if numPart != 0 {
			numPart--
		}

		newToken := NewBytePair(token.First(), numPart)
		rules := l.ByteRules[newToken]
		if rules.Weights == nil {
			output.Append(newToken)
			continue
		}

		output.AppendSlice(rules.ChooseSuccessor())
	}
}

func (l *LSystem) IterateUntil(n int) []BytePair {
	l.Reset()
	if n >= 5 {
		n -= 5
		l.prime(5)
	}
	l.applyRules(n)
	return l.MemPool.ReadAll()
}

func (l *LSystem) prime(n int) {
	for i := 0; i < n; i++ {
		l.applyRulesOnce(l.MemPool.GetReadBuffer(0), l.MemPool.GetWriteBuffer(0))
		l.MemPool.Swap(0)
	}

	chunkSize := l.MemPool.GetReadBuffer(0).Len / threadCount
	for i := 0; i < threadCount; i++ {
		from := i * chunkSize
		to := from + chunkSize
		if i == threadCount-1 {
			to = l.MemPool.GetReadBuffer(0).Len
		}

		l.MemPool.GetWriteBuffer(i).AppendSlice(l.MemPool.GetReadBuffer(0).BytePairs[from:to])
	}
	for i := 0; i < threadCount; i++ {
		l.MemPool.Swap(i)
	}
}

func (l *LSystem) Iterate(n int) []BytePair {
	l.applyRules(n)

	return l.MemPool.ReadAll()
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
	l.MemPool.Reset()
	l.MemPool.GetReadBuffer(0).Append(l.TokenBytes[l.Axiom])
	l.MemPool.GetReadBuffer(0).Len = 1
}
