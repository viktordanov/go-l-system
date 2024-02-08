package lsystem

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var benchmarkRules = map[Token]string{
	"Seed": `1 L u S2`,
	"L":    `0.1 L u L w F e; 0.1 L_ u L e F w; 0.1 L_ u L n F s; 0.1 L_ u L s F n; 0.04 L_ [ w L_ w u seed ]; 0.04 L_ [ e L_ e u seed ]; 0.04 L_ [ s L_ s u seed ]; 0.04 L_ [ n L_ n u seed ]; 0.05 L_ u L; 1 L`,
	"S2":   `1 [ n F ] [ w F ] [ s F ] [ e F ] u n S1; 1 [ n F ] [ w F ] [ s F ] [ e F ] u w S1; 1 [ n F ] [ w F ] [ s F ] [ e F ] u s S1; 1 [ n F ] [ w F ] [ s F ] [ e F ] u e S1`,
	"S1":   `1 [ n F ] [ w F ] [ s F ] [ e F ] u n S0; 1 [ n F ] [ w F ] [ s F ] [ e F ] u w S0; 1 [ n F ] [ w F ] [ s F ] [ e F ] u s S0; 1 [ n F ] [ w F ] [ s F ] [ e F ] u e S0`,
	"S0":   `1 [ n F ] [ w F ] [ s F ] [ e F ] n S0; 1 [ n F ] [ w F ] [ s F ] [ e F ] w S0; 1 [ n F ] [ w F ] [ s F ] [ e F ] s S0; 1 [ n F ] [ w F ] [ s F ] [ e F ] e S0`,
	"F":    `0.005 F [ d D ]; 0.001 F [ u F_ ]; 0.0O8 F [ n F_ ]; 0.0O8 F [ w F_ ]; 0.0O8 F [ e F_ ]; 0.0O8 F [ s F_ ]; 1 F`,
}

func BenchmarkParseRules(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseRules(benchmarkRules)
	}
}

func BenchmarkLSystemIterateAB(b *testing.B) {
	rulesStr := map[Token]string{
		"A": "1 A B",
		"B": "1 A",
	}

	vars, consts, rules := ParseRules(rulesStr)
	lsys := NewLSystem("A", rules, vars, consts, true)

	tests := []struct {
		name  string
		iters int
	}{
		{"10", 10},
		{"20", 20},
		//{"30", 30},
		//{"40", 40},
	}

	b.ResetTimer()
	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				lsys.IterateUntil(tt.iters)
			}
		})
	}
}

func BenchmarkLSystemIterate(b *testing.B) {
	vars, consts, rules := ParseRules(benchmarkRules)
	ls := NewLSystem("Seed", rules, vars, consts, true)
	tests := []struct {
		name  string
		iters int
	}{
		{"10", 10},
		{"20", 20},
		{"30", 30},
		{"100", 100},
	}

	b.ResetTimer()
	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				ls.IterateUntil(tt.iters)
			}
		})
	}
}

func BenchmarkChooseSuccessor(b *testing.B) {
	r := NewProductionRule("L", ParseRule(`0.1 L u L w F e; 0.1 L_ u L e F w; 0.1 L_ u L n F s; 0.1 L_ u L s F n; 0.04 L_ [ w L_ w u seed ]; 0.04 L_ [ e L_ e u seed ]; 0.04 L_ [ s L_ s u seed ]; 0.04 L_ [ n L_ n u seed ]; 0.05 L_ u L; 1 L`))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ChooseSuccessor()
	}
}

func BenchmarkChooseSuccessorBytes(b *testing.B) {
	vars, consts, rules := ParseRules(benchmarkRules)
	ls := NewLSystem("Seed", rules, vars, consts, true)
	r := NewProductionRule("L", ParseRule(`0.1 L u L w F e; 0.1 L_ u L e F w; 0.1 L_ u L n F s; 0.1 L_ u L s F n; 0.04 L_ [ w L_ w u seed ]; 0.04 L_ [ e L_ e u seed ]; 0.04 L_ [ s L_ s u seed ]; 0.04 L_ [ n L_ n u seed ]; 0.05 L_ u L; 1 L`))

	br := ls.ByteRules[ls.TokenBytes[r.Predecessor]]

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		br.ChooseSuccessor(ls, ls.EmptyTokenId)
	}
}

func BenchmarkByteTokenPacking(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewTokenStateId(121, true)
	}
}

func BenchmarkByteTokenUnpacking(b *testing.B) {
	t := NewTokenStateId(121, true)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.TokenId()
		t.HasParam()
	}
}

func TestCounterVariables(t *testing.T) {
	var counterRules = map[Token]string{
		"Seed": `1 L u S3`,
		"L":    `1 L`,
		"S1":   `1 X`,
	}
	vars, consts, rules := ParseRules(counterRules)
	ls := NewLSystem("Seed", rules, vars, consts, true)

	ls.IterateOnce()
	assertState(t, []Token{"L", "u", "S3"}, ls.DecodeBytes(ls.MemPool.ReadAll()))

	ls.IterateOnce()
	assertState(t, []Token{"L", "u", "S2"}, ls.DecodeBytes(ls.MemPool.ReadAll()))

	ls.IterateOnce()
	assertState(t, []Token{"L", "u", "X"}, ls.DecodeBytes(ls.MemPool.ReadAll()))
}

func TestCatalystParse(t *testing.T) {
	var catalystRules = map[Token]string{
		"A": `1 A B`,
		"B": `1 *B A A`,
	}
	vars, consts, rules := ParseRules(catalystRules)
	ls := NewLSystem("Seed", rules, vars, consts, false)

	ls.IterateOnce()
	assertState(t, []Token{"A", "B"}, ls.DecodeBytes(ls.MemPool.ReadAll()))
	ls.IterateOnce()
	assertState(t, []Token{"A", "B", "B"}, ls.DecodeBytes(ls.MemPool.ReadAll()))
	ls.IterateOnce()
	assertState(t, []Token{"A", "B", "B", "A", "A"}, ls.DecodeBytes(ls.MemPool.ReadAll()))
	ls.IterateOnce()
	assertState(t, []Token{"A", "B", "B", "A", "A", "A", "B", "A", "B"}, ls.DecodeBytes(ls.MemPool.ReadAll()))
}

func assertState(t *testing.T, expected, actual []Token) {
	assert.Equal(t, len(expected), len(actual))
	assert.EqualValues(t, expected, actual)
}
