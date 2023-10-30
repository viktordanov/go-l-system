package lsystem

import (
	"strconv"
	"strings"
	"sync"
)

type LSystem struct {
	Axiom                Token
	Rules                map[Token]ProductionRule
	Variables            TokenSet
	Constants            TokenSet
	useWeightPresampling bool

	TokenBytes map[Token]TokenStateId
	BytesToken [255]Token
	ByteRules  [255]ByteProductionRule
	Params     [128]uint8
	MemPool    *MemPool
}

func NewLSystem(axiom Token, rulesMap map[Token]ProductionRule, vars TokenSet, consts TokenSet, useWeightPresampling bool) *LSystem {
	lSystem := &LSystem{
		Axiom:     axiom,
		Rules:     rulesMap,
		Variables: vars,
		Constants: consts,
		MemPool:   NewMemPool(32),

		useWeightPresampling: useWeightPresampling,
	}

	lSystem.encodeTokens()
	lSystem.Reset()
	return lSystem
}

func (l *LSystem) Recreate(byteRules [255]ByteProductionRule) *LSystem {
	clone := *l
	clone.ByteRules = byteRules
	return &clone
}

func (l *LSystem) RecreateWithMemPool(byteRules [255]ByteProductionRule, pool *MemPool) *LSystem {
	clone := *l
	clone.ByteRules = byteRules
	clone.MemPool = pool
	return &clone
}

func (l *LSystem) encodeTokens() {
	l.TokenBytes = make(map[Token]TokenStateId)
	l.BytesToken = [255]Token{}
	i := uint8(0)

	statefulVarParams := make(map[Token]uint8)
	for t := range l.Variables {
		baseVar, numberState, isStateful := tryParseStatefulVariable(t)
		if isStateful {
			baseVar := Token(baseVar)
			if _, exists := statefulVarParams[baseVar]; !exists {
				statefulVarParams[baseVar] = numberState
			}
			statefulVarParams[baseVar] = max(numberState, statefulVarParams[baseVar])
		} else {
			bytePair := NewTokenStateId(i, false)
			l.TokenBytes[t] = bytePair
			l.BytesToken[bytePair] = t
			i++
		}
	}

	for t := range l.Constants {
		bytePair := NewTokenStateId(i, false)
		l.TokenBytes[t] = bytePair
		l.BytesToken[bytePair] = t
		i++
	}

	j := 0
	for baseVar, maxState := range statefulVarParams {
		minIndex := 1
		maxIndex := int(maxState)
		for k := minIndex; k <= maxIndex; k++ {
			bytePair := NewTokenStateId(uint8(j), true)
			l.TokenBytes[Token(string(baseVar)+strconv.Itoa(k))] = bytePair
			l.BytesToken[bytePair] = Token(string(baseVar) + strconv.Itoa(k))
			l.Params[j] = uint8(k)
			j++
		}
	}

	l.ByteRules = [255]ByteProductionRule{}
	for t, rule := range l.Rules {
		l.ByteRules[l.TokenBytes[t]] = rule.EncodeTokens(l.TokenBytes, l.useWeightPresampling)
	}
}

func (l *LSystem) EncodeTokens(tokens []Token) []TokenStateId {
	result := make([]TokenStateId, 0, len(tokens))
	for _, t := range tokens {
		result = append(result, l.TokenBytes[t])
	}
	return result
}

func (l *LSystem) DecodeBytes(bp []TokenStateId) []Token {
	result := make([]Token, 0, len(bp))
	for _, bytePair := range bp {
		//v, exists := l.BytesToken[bytePair]
		v := l.BytesToken[bytePair]
		exists := v != ""
		if !exists {
			base := bytePair.TokenId()
			v = l.BytesToken[NewTokenStateId(base, false)]
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
		if token.HasParam() && l.Params[token.TokenId()] > 1 {
			token--
		}
		rules := l.ByteRules[token]
		if rules.Weights == nil {
			output.Append(token)
			continue
		}

		output.AppendSlice(rules.ChooseSuccessor())
		l.ByteRules[token] = rules
	}
}

func (l *LSystem) IterateUntil(n int) []TokenStateId {
	l.Reset()
	if n >= 15 {
		n -= 10
		l.prime(10)
		l.applyRules(n)
	} else {
		for i := 0; i < n; i++ {
			l.applyRulesOnce(l.MemPool.GetReadBuffer(0), l.MemPool.GetWriteBuffer(0))
			l.MemPool.Swap(0)
		}
	}
	return l.MemPool.ReadAll()
}

func (l *LSystem) prime(n int) {
	for i := 0; i < n; i++ {
		l.applyRulesOnce(l.MemPool.GetReadBuffer(0), l.MemPool.GetWriteBuffer(0))
		l.MemPool.Swap(0)
	}

	l.distribute()
}

func (l *LSystem) distribute() {
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

func (l *LSystem) Iterate(n int) []TokenStateId {
	l.applyRules(n)

	return l.MemPool.ReadAll()
}

func (l *LSystem) IterateOnce() []TokenStateId {
	l.applyRulesOnce(l.MemPool.GetReadBuffer(0), l.MemPool.GetWriteBuffer(0))
	l.MemPool.Swap(0)

	buffer := l.MemPool.GetReadBuffer(0)
	return buffer.BytePairs[:buffer.Len]
}

func (l *LSystem) String() string {
	var sb strings.Builder
	for tokenId, rule := range l.ByteRules {
		if rule.Weights == nil {
			continue
		}
		sb.WriteString("\"" + string(l.BytesToken[tokenId]) + "\": ")
		sb.WriteString(rule.String(l.BytesToken))
		sb.WriteString(",\n")
	}
	return sb.String()
}

func (l *LSystem) Reset() {
	l.MemPool.Reset()
	l.MemPool.GetReadBuffer(0).Append(l.TokenBytes[l.Axiom])
	l.MemPool.GetReadBuffer(0).Len = 1
}
