package express

import (
	"log"
	"math/rand"
	"strings"
	"unicode/utf8"
)

var (
	ALPHA      = strings.Split("abcdefghijklmnopqrstuvwxyz", "")
	ALPHA_CAP  = strings.Split("ABCDEFGHIJKLMNOPQRSTUVWXYZ", "")
	NUMERICAL  = strings.Split("0123456789", "")
	WHITESPACE = strings.Split("\t\n\f\r", "")
)

func Find(list []string, target string) int {
	for i, e := range list {
		if e == target {
			return i
		}
	}
	return -1
}

type Exp interface {
	Get() string
}

type FixedExp struct {
	value string
}

func NewFixedExp(value string) FixedExp {
	return FixedExp{
		value: value,
	}
}

func (exp FixedExp) Get() string {
	return exp.value
}

type OrExp struct {
	first  Exp
	second Exp
}

func NewOrExp(first Exp, second Exp) OrExp {
	return OrExp{
		first:  first,
		second: second,
	}
}

func (exp OrExp) Get() string {
	flip := rand.Float32()

	if flip > 0.5 {
		return exp.first.Get()
	} else {
		return exp.second.Get()
	}
}

type RepExpZero struct {
	value Exp
	max   int
}

func NewRepExpZero(value Exp, max int) RepExpZero {
	return RepExpZero{
		value: value,
		max:   max,
	}
}

func (exp RepExpZero) Get() string {
	n := rand.Intn(exp.max + 1)

	res := make([]string, n)
	for i := 0; i < n; i++ {
		res[i] = exp.value.Get()
	}

	return strings.Join(res, "")
}

type RepExpOne struct {
	value Exp
	max   int
}

func NewRepExpOne(value Exp, max int) RepExpOne {
	return RepExpOne{
		value: value,
		max:   max,
	}
}

func (exp RepExpOne) Get() string {
	n := rand.Intn(exp.max) + 1

	res := make([]string, n)
	for i := 0; i < n; i++ {
		res[i] = exp.value.Get()
	}

	return strings.Join(res, "")
}

type RepExpFixed struct {
	value Exp
	count int
}

func NewRepExpFixed(value Exp, count int) RepExpFixed {
	return RepExpFixed{
		value: value,
		count: count,
	}
}

func (exp RepExpFixed) Get() string {
	res := make([]string, exp.count)
	for i := 0; i < exp.count; i++ {
		res[i] = exp.value.Get()
	}

	return strings.Join(res, "")
}

type ExpOptional struct {
	value Exp
}

func NewExpOptional(value Exp) ExpOptional {
	return ExpOptional{
		value: value,
	}
}

func (exp ExpOptional) Get() string {
	flip := rand.Float32()
	if flip > 0.5 {
		return ""
	} else {
		return exp.value.Get()
	}
}

type ExpAny struct{}

func NewExpAny() ExpAny {
	return ExpAny{}
}

func (exp ExpAny) Get() string {
	n := rand.Intn(4) + 1

	var bytes []byte = make([]byte, 0, n)
	if n == 1 {
		bytes[1] = byte(uint8(rand.Intn(128)))
	} else {
		if n == 2 {
			bytes[1] = byte(uint8(rand.Intn(30) + 194))
		} else if n == 3 {
			bytes[1] = byte(uint8(rand.Intn(16) + 223))
		} else {
			bytes[1] = byte(uint8(rand.Intn(8) + 240))
		}

		for i := 1; i < n; i++ {
			bytes[i] = byte(uint8(rand.Intn(64) + 128))
		}
	}

	char, _ := utf8.DecodeRune(bytes)
	return string(char)
}

type ExpAnyDigit struct{}

func NewExpAnyDigit() ExpAnyDigit {
	return ExpAnyDigit{}
}

func (exp ExpAnyDigit) Get() string {
	return NUMERICAL[rand.Intn(len(NUMERICAL))]
}

type ExpAlphaRange struct {
	rang   int
	offset int
}

func NewExpAlphaRange(low string, high string) ExpAlphaRange {
	hi, lo := Find(ALPHA, high), Find(ALPHA, low)

	if hi == -1 || lo == -1 || len(low) > 1 || len(high) > 1 {
		log.Panicf("Received invalid characters for alphabetical range expression low=%s, high=%s. Expects single alphabetical characters.", low, high)
	}

	return ExpAlphaRange{
		rang:   hi - lo,
		offset: lo,
	}
}

func (exp ExpAlphaRange) Get() string {
	return ALPHA[rand.Intn(exp.rang)+exp.offset]
}

type ExpAlphaCapRange struct {
	rang   int
	offset int
}

func NewExpAlphaCapRange(low string, high string) ExpAlphaCapRange {
	hi, lo := Find(ALPHA_CAP, high), Find(ALPHA_CAP, low)

	if hi == -1 || lo == -1 || len(low) > 1 || len(high) > 1 {
		log.Panicf("Received invalid characters for alphabetical range expression low=%s, high=%s. Expects single alphabetical characters.", low, high)
	}

	return ExpAlphaCapRange{
		rang:   hi - lo,
		offset: lo,
	}
}

func (exp ExpAlphaCapRange) Get() string {
	return ALPHA_CAP[rand.Intn(exp.rang)+exp.offset]
}

type ExpDigitRange struct {
	rang   int
	offset int
}

func NewExpDigitRange(low string, high string) ExpDigitRange {
	hi, lo := Find(NUMERICAL, high), Find(NUMERICAL, low)

	if hi == -1 || lo == -1 || len(low) > 1 || len(high) > 1 {
		log.Panicf("Received invalid characters for numerical range expression low=%s, high=%s. Expects single numeric characters.", low, high)
	}

	return ExpDigitRange{
		rang:   hi - lo,
		offset: lo,
	}
}

func (exp ExpDigitRange) Get() string {
	return NUMERICAL[rand.Intn(exp.rang)+exp.offset]
}

type ExpVariable struct {
	key    string
	parent map[string]string
	source *Exp
}

func NewExpVariable(key string, parent map[string]string, source *Exp) ExpVariable {
	return ExpVariable{
		key:    key,
		parent: parent,
		source: source,
	}
}

func (exp ExpVariable) Get() string {
	if exp.source != nil {
		exp.parent[exp.key] = (*exp.source).Get()
	}

	return exp.parent[exp.key]
}
