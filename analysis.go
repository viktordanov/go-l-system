package lsystem

import (
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
)

// AnalyseProductionRates analyses the production rate of each production rule
// by creating a new LSystem for each rule containing only its own rules
// and itself as axiom. Then it iterates the LSystem for a given number of
// iterations and analyzing the distribution of the produced tokens.
func (l *LSystem) AnalyseProductionRates() map[Token]ProductionRate {
	tokensProduced := make(map[Token]ProductionRate)
	const iterations = 1

	fnCalcAvgGrowth := func(rate ProductionRate) {
		totalSum := 0
		avgGrowingRate := 0.0
		for _, rate := range rate.Rates {
			totalSum += int(rate)
		}

		for idx, rate := range rate.Rates {
			avgGrowingRate += float64(idx) * float64(rate) / float64(totalSum)
		}

		log.Println("Production Rates (Avg growth " + strconv.FormatFloat(avgGrowingRate/1000, 'f', 4, 64) + ")")
	}

	fnSampleLsystem := func(token Token, lsystem *LSystem, iters, samples int) {
		for i := 0; i < iters; i++ {
			lsystem.Reset()
			prevLen := 0
			for j := 0; j < samples; j++ {
				tokens := lsystem.IterateOnce()
				length := len(tokens)
				if prevLen == 0 {
					prevLen = length
					continue
				}
				diff := int(math.Round(float64(length) / float64(prevLen) * 1000))
				if diff >= len(tokensProduced[token].Rates) {
					replacement := make([]float32, diff*2)
					copy(replacement, tokensProduced[token].Rates)
					tokensProduced[token] = ProductionRate{
						Token: token,
						Rates: replacement,
						Rule:  tokensProduced[token].Rule,
					}
				}
				prevLen = length
				tokensProduced[token].Rates[diff]++
			}

			fnCalcAvgGrowth(tokensProduced[token])
			delete(tokensProduced, token)
		}
	}

	//for token, rule := range l.Rules {
	//	r := rule
	//	tokensProduced[token] = ProductionRate{
	//		Token: token,
	//		Rates: make([]float32, 1024),
	//		Rule:  &r,
	//	}
	//	lsystem := NewLSystem(token, map[Token]ProductionRule{token: r}, l.Variables, l.Constants, false)
	//	fnSampleLsystem(token, lsystem, iterations, 10)
	//}

	tokensProduced["LSystem"] = ProductionRate{
		Token: "LSystem",
		Rates: make([]float32, 1024),
		Rule:  l,
	}
	fnSampleLsystem("LSystem", l, iterations, 80)

	return tokensProduced
}

type httpHandler func(w http.ResponseWriter, r *http.Request)

func (l *LSystem) generateChartHandlers() map[Token]httpHandler {
	productionRates := l.AnalyseProductionRates()
	handlers := make(map[Token]httpHandler)
	for _, rate := range productionRates {
		r := rate
		handlers[rate.Token] = func(w http.ResponseWriter, _ *http.Request) {
			if err := r.RenderChart(w); err != nil {
				panic(err)
			}
		}
	}
	return handlers
}

func (l *LSystem) HandleStatisticsServer(w http.ResponseWriter, _ *http.Request) {
	productionRates := l.AnalyseProductionRates()

	for _, rate := range productionRates {
		if err := rate.RenderChart(w); err != nil {
			panic(err)
		}

	}
}

func (pr *ProductionRate) RenderChart(w io.Writer) error {
	// create a new bar instance
	bar := charts.NewBar()
	// set some global options like Title/Legend/ToolTip or anything else
	bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    "Production Rate Analysis",
		Subtitle: "Analysis of production rates of the production rule " + string(pr.Token) + " with " + strconv.Itoa(len(pr.Rates)) + " iterations " + pr.Rule.String(),
	}))

	// Convert production rates into bar items
	barItems := make([]opts.BarData, len(pr.Rates), len(pr.Rates))
	lastNonZero := 0

	totalSum := 0
	avgGrowingRate := 0.0
	for _, rate := range pr.Rates {
		totalSum += int(rate)
	}

	for idx, rate := range pr.Rates {
		if rate > 0 {
			lastNonZero = idx
		}
		barItems[idx] = opts.BarData{Value: rate}
		avgGrowingRate += float64(idx) * float64(rate) / float64(totalSum)
	}

	labelsUntilLast := make([]string, lastNonZero+1, lastNonZero+1)
	for i := 0; i < lastNonZero+1; i++ {
		labelsUntilLast[i] = strconv.Itoa(i)
	}

	title := "Production Rates (Avg growth " + strconv.FormatFloat(avgGrowingRate/1000, 'f', 4, 64) + ")"

	// Put data into instance
	bar.SetXAxis(labelsUntilLast).
		AddSeries(title, barItems[0:lastNonZero+1])
	return bar.Render(w)
}

func (l *LSystem) Serve() error {
	for token, handler := range l.generateChartHandlers() {

		log.Println("Registering handler for token", token)
		http.HandleFunc("/"+string(token), handler)
	}
	return http.ListenAndServe(":8081", nil)
}
