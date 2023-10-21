package main

import (
	. "github.com/viktordanov/lsystem"
)

func main() {
	rulesStr := map[Token]string{
		"Seed": `1 L u S2`,
		"L":    `0.1 L u L w F e; 0.1 L_ u L e F w; 0.1 L_ u L n F s; 0.1 L_ u L s F n; 0.04 L_ [ w L_ w u seed ]; 0.04 L_ [ e L_ e u seed ]; 0.04 L_ [ s L_ s u seed ]; 0.04 L_ [ n L_ n u seed ]; 0.05 L_ u L; 1 L`,
		"S2":   `1 [ n F ] [ w F ] [ s F ] [ e F ] u n S1; 1 [ n F ] [ w F ] [ s F ] [ e F ] u w S1; 1 [ n F ] [ w F ] [ s F ] [ e F ] u s S1; 1 [ n F ] [ w F ] [ s F ] [ e F ] u e S1`,
		"S1":   `1 [ n F ] [ w F ] [ s F ] [ e F ] u n S0; 1 [ n F ] [ w F ] [ s F ] [ e F ] u w S0; 1 [ n F ] [ w F ] [ s F ] [ e F ] u s S0; 1 [ n F ] [ w F ] [ s F ] [ e F ] u e S0`,
		"S0":   `1 [ n F ] [ w F ] [ s F ] [ e F ] n S0; 1 [ n F ] [ w F ] [ s F ] [ e F ] w S0; 1 [ n F ] [ w F ] [ s F ] [ e F ] s S0; 1 [ n F ] [ w F ] [ s F ] [ e F ] e S0`,
		"F":    `0.005 F [ d D ]; 0.001 F [ u F_ ]; 0.0O8 F [ n F_ ]; 0.0O8 F [ w F_ ]; 0.0O8 F [ e F_ ]; 0.0O8 F [ s F_ ]; 1 F`,
	}

	vars, consts, rules := ParseRules(rulesStr)
	lsys := NewLSystem("Seed", rules, vars, consts)

	for i := 0; i < 60; i++ {
		lsys.Iterate(i)
	}
}
