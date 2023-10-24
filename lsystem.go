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
	MemPool    *BufferPool
}

func NewLSystem(axiom Token, rulesMap map[Token]ProductionRule, vars TokenSet, consts TokenSet) *LSystem {
	lSystem := &LSystem{
		Axiom:     axiom,
		Rules:     rulesMap,
		Variables: vars,
		Constants: consts,
		MemPool:   NewBufferPool(1024),
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

func (l *LSystem) applyRules() {
	input := l.MemPool.GetSwap()
	n := len(input.BytePairs[0:input.Len])
	numGoroutines := 4
	chunkSize := n / numGoroutines

	var wg sync.WaitGroup
	// Use a slice of slices to store results
	results := make([][]BytePair, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if i == numGoroutines-1 {
			end = n
		}

		wg.Add(1)
		go func(idx, start, end int) {
			defer wg.Done()
			var output []BytePair
			for _, token := range input.BytePairs[start:end] {
				numPart := token.Second()
				if numPart != 0 {
					numPart--
				}
				newToken := NewBytePair(token.First(), numPart)
				rules := l.ByteRules[newToken]
				if rules.Weights == nil {
					output = append(output, newToken)
				} else {
					output = append(output, rules.ChooseSuccessor()...)
				}
			}
			results[idx] = output
		}(i, start, end)
	}

	wg.Wait()

	// Collect results in order
	for i := 0; i < numGoroutines; i++ {
		l.MemPool.AppendSlice(results[i])
	}

	l.MemPool.Swap()
	l.MemPool.ResetWritingHead()
}

func (l *LSystem) IterateUntil(n int) []BytePair {
	l.Reset()
	for i := 0; i < n; i++ {
		l.applyRules()
	}
	return l.MemPool.GetSwap().BytePairs[:l.MemPool.GetSwap().Len]
}

func (l *LSystem) Iterate(n int) []BytePair {
	for i := 0; i < n; i++ {
		l.applyRules()
	}

	return l.MemPool.GetSwap().BytePairs[:l.MemPool.GetSwap().Len]
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
	l.MemPool.GetSwap().BytePairs[0] = l.TokenBytes[l.Axiom]
	l.MemPool.GetSwap().Len = 1
}
