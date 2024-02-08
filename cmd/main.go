package main

import (
	"flag"
	. "github.com/viktordanov/lsystem"
	"log"
	"os"
	"runtime/pprof"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	var catalystRules = map[Token]string{
		"S":  `1 C9 A`,
		"C4": `1 _`,
		"A":  `1 *C A A; 1 *C A`,
	}
	vars, consts, rules := ParseRules(catalystRules)
	ls := NewLSystem("S", rules, vars, consts, false)

	for i := 0; i < 10; i++ {
		ls.IterateOnce()
		log.Println(ls.DecodeBytes(ls.MemPool.ReadAll()))
	}
}
