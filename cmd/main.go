package main

import (
	"flag"
	"fmt"
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
		"A":  "1 A; 1 A B C6",
		"C5": "1 x",
		"B":  "1 o; 1 x",
	}

	vars, consts, rules := ParseRules(rulesStr)
	lsys := NewLSystem("A", rules, vars, consts, true)

	lsys.Reset()
	fmt.Println(lsys.DecodeBytes(lsys.MemPool.ReadAll()))
	for i := 0; i < 10; i++ {
		fmt.Println(lsys.DecodeBytes(lsys.IterateOnce()))
	}

	fmt.Println(lsys)
}
