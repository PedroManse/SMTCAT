package service

import (
	"regexp"
	"os"
	"io"
	"fmt"
	"bytes"
	"gopkg.in/Knetic/govaluate.v2"
	"strconv"
)


/*

store file manual

group selection: https://regexr.com/7ovk0
input option:    https://regexr.com/7ovkc
input range:     https://regexr.com/7ovkf

Knetic's GoValuate repo is used to perform math operations with the prices
https://github.com/Knetic/govaluate
*/


func MustReadFile(filename string) []byte {
	file, e := os.Open(filename)
	if (e!=nil) {panic(e)}
	defer file.Close()
	FILE, e := io.ReadAll(file)
	if (e!=nil) {panic(e)}
	return FILE
}

type StageMode = int
const (
	ModeRange StageMode = iota
	ModeOption
)

func cartEval(expr, input string) string {
	expression, err := govaluate.NewEvaluableExpression(expr);
	if (err != nil) {panic(err)}
	parameters := map[string]any{
		"input":input,
	}
	result, err := expression.Evaluate(parameters)
	if (err != nil) {panic(err)}
	return result.(string)
}

func optpriceEval(expr string, price float64, input string) (newprice float64) {
	expression, err := govaluate.NewEvaluableExpression(expr)
	if (err != nil) {panic(err)}
	parameters := map[string]any{
		"price":price,
		"input":input,
	}
	result, err := expression.Evaluate(parameters)
	if (err != nil) {panic(err)}
	return result.(float64)
}

func rngpriceEval(expr string, price float64, input int64) (newprice float64) {
	expression, err := govaluate.NewEvaluableExpression(expr);
	if (err != nil) {panic(err)}
	parameters := map[string]any{
		"price":price,
		"input":input,
	}
	result, err := expression.Evaluate(parameters)
	if (err != nil) {panic(err)}
	return result.(float64)
}

func (opts Ooption) PriceEval(price float64, input string) (newprice float64) {
	for _, opt := range opts.Options {
		if (opt.text == input) {
			if (opt.priceExpr == "") {return price}
			return optpriceEval(opt.priceExpr, price, input)
		}
	}
	// assert false
	return -1
}

func (rng Orange) PriceEval(price float64, input string) (newprice float64) {
	if (rng.priceExpr == "") {return price}
	v, e := strconv.ParseInt(input, 10, 64)
	if (e != nil) {panic(e)}
	return rngpriceEval(rng.priceExpr, price, v)
}

func (opts Ooption) CartEval(input string) (string) {
	for _, opt := range opts.Options {
		if (opt.text == input) {
			return cartEval(opt.cartExpr, input)
		}
	}
	// assert false
	return fmt.Sprintf("ERROR no option [%q]", input)
}

func (rng Orange) CartEval(input string) (string) {
	return cartEval(rng.cartExpr, input)
}

func (opts Ooption) Validate(input string) (ok bool) {
	for _, opt := range opts.Options {
		ok = ok || (opt.text == input)
	}
	return ok
}

func (opts Ooption) NextOptName(input string) (string) {
	for _, opt := range opts.Options {
		if (opt.text == input) {
			return opt.nextOpt
		}
	}
	// assert false
	return ""
}

func (rng Orange) Validate(input string) (bool) {
	v, e := strconv.ParseInt(input, 10, 64)
	return e==nil || (rng.min <= v && v <= rng.max)
}

func (rng Orange) NextOptName(input string) (string) {
	return rng.nextOpt
}

type CartOption interface {
	// usable input
	Validate(input string) (ok bool)
	// what to show in cart
	CartEval(input string) string
	// how to modify price
	PriceEval(price float64, input string) (newprice float64)
	// next option
	NextOptName(input string) string
}

type _option struct {
	text string
	cartExpr string
	priceExpr string
	nextOpt string
}

type Ooption struct {
	Options []_option
}

type Orange struct {
	min int64
	max int64
	cartExpr string
	nextOpt string
	priceExpr string
}

type Stage struct {
	StageName []byte
	StageText []byte
	Opt CartOption
	MayEnd bool
}

func init() {
	StageFinder := regexp.MustCompile(`@([\w ]+)\n"([\w() ]+)"\n((?:.|\n.)+?)\n\n`)
	OptionFinder := regexp.MustCompile(`^"([\w() ]+)":(\(.*?\))->([\w ]+?) ?(\(.*\))?$`)
	RangeFinder := regexp.MustCompile(`^"(\d+)-(\d+)":(\(.*?\))->([\w ]+?) ?(\(.*\))?$`)

	StoreFile := MustReadFile("qs.txt")
	stages := StageFinder.FindAllSubmatch(StoreFile, -1)
	for _, stage := range stages {
		if len(stage) < 4 {
			panic(fmt.Errorf("`%s`\n doesn't have [group name, group text, options]", stage))
		}
		//StageText := stage[0]
		StageName := stage[1]
		StageText := stage[2]
		StageOptions := bytes.Split(stage[3], []byte("\n"))
		if len(StageOptions[len(StageOptions)-1]) == 0 {
			StageOptions = StageOptions[:len(StageOptions)-1]
		}

		var RangeOrOption StageMode
		var StageOption CartOption
		var opbf []_option = make([]_option, len(StageOptions))
		for i, option := range StageOptions {
			opt := OptionFinder.FindStringSubmatch(string(option))
			rng := RangeFinder.FindStringSubmatch(string(option))
			if (opt != nil) {
				RangeOrOption = ModeOption
				opbf[i] = _option{
					opt[1],
					opt[2],
					opt[3],
					opt[4],
				}
			} else if (rng != nil) {
				RangeOrOption = ModeRange
				// assert len(StageOptions) == 1
				vmin, e := strconv.ParseInt(rng[1], 10, 64)
				if (e!=nil) {panic(e)}
				vmax, e := strconv.ParseInt(rng[2], 10, 64)
				if (e!=nil) {panic(e)}
				StageOption = Orange{
					min: vmin,
					max: vmax,
					cartExpr: rng[3],
					nextOpt: rng[4],
					priceExpr: rng[5],
				}
				break
			} else {
				panic(fmt.Errorf("`%s` doens't fit neither in the option or range regex", option))
			}
		}
		if (RangeOrOption == ModeOption) {
			StageOption = Ooption{ opbf }
		}

		fmt.Println("is range:", RangeOrOption == ModeRange)
		fmt.Println(StageOption)
		//Stage{RangeName, StageText}

		fmt.Printf("Name: %s\nText: %q\nOptions: %s\n",
		StageName, StageText, StageOptions)
		fmt.Println("==============================")
	}
}
