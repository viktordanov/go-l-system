package main

import (
	"flag"
	"log"
	"math/rand"
	"os"
	"runtime/pprof"

	. "github.com/viktordanov/lsystem"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	// seed random number generator
	rand.Seed(1)

	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	rulesStr := map[Token]string{
		"A": "1 A B",
		"B": "1 A",
	}

	vars, consts, rules := ParseRules(rulesStr)
	lsys := NewLSystem("A", rules, vars, consts)

	for i := 0; i < 100; i++ {
		lsys.Reset()
		lsys.IterateUntil(30)
	}
}
