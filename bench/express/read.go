package express

import (
	"log"
	"math/rand"
	"strconv"
	"strings"
	"unicode"
)

type ExpGroup struct {
	list []Exp
	vars map[string]string
}

func NewExpGroup(parent map[string]string) ExpGroup {
	if parent != nil {
		return ExpGroup{
			list: make([]Exp, 0),
			vars: parent,
		}
	} else {
		return ExpGroup{
			list: make([]Exp, 0),
			vars: make(map[string]string),
		}
	}
}

func (exp ExpGroup) Get() string {
	res := make([]string, len(exp.list))
	for i, value := range exp.list {
		res[i] = value.Get()
	}

	return strings.Join(res, "")
}

func (exp *ExpGroup) AddExp(newExp Exp) {
	exp.list = append(exp.list, newExp)
}

func (exp *ExpGroup) PopExp() Exp {
	last := exp.list[len(exp.list)-1]
	exp.list = exp.list[:len(exp.list)-1]

	return last
}

func (exp ExpGroup) Len() int {
	return len(exp.list)
}

type OrExpGroup struct {
	list []Exp
}

func NewOrExpGroup() OrExpGroup {
	return OrExpGroup{
		list: make([]Exp, 0),
	}
}

func (exp OrExpGroup) Get() string {
	return exp.list[rand.Intn(len(exp.list))].Get()
}

func (exp *OrExpGroup) AddExp(newExp Exp) {
	exp.list = append(exp.list, newExp)
}

func (exp *OrExpGroup) PopExp() Exp {
	last := exp.list[len(exp.list)-1]
	exp.list = exp.list[:len(exp.list)-1]

	return last
}

func (exp OrExpGroup) Len() int {
	return len(exp.list)
}

func parseInt(regex []rune) int {
	x, err := strconv.Atoi(string(regex))
	if err != nil {
		log.Fatal(err)
	}

	return x
}

func parseClass(regex []rune) Exp {
	res, ctr := NewOrExpGroup(), 0

	var c1, c2 rune
	for ctr < len(regex) {
		c1, c2 = regex[ctr], regex[ctr+2]

		if unicode.IsLetter(c1) && unicode.IsLetter(c2) {
			if unicode.IsUpper(c1) && unicode.IsUpper(c2) {
				res.AddExp(NewExpAlphaCapRange(string(c1), string(c2)))
			} else if unicode.IsLower(c1) && unicode.IsLower(c2) {
				res.AddExp(NewExpAlphaRange(string(c1), string(c2)))
			} else {
				panic("Found invalid character class range. Capital and non-capital letter!")
			}
		} else if unicode.IsDigit(c1) && unicode.IsNumber(c2) {
			res.AddExp(NewExpDigitRange(string(c1), string(c2)))
		} else {
			panic("Found invalid character class range. Numeric and non-numeric!")
		}

		ctr += 3
	}

	if res.Len() == 1 {
		return res.PopExp()
	}

	return res
}

func parseGroup(regex []rune, parent map[string]string) Exp {
	res, ctr := NewExpGroup(parent), 0

	var char rune
	for ctr < len(regex) {
		char = regex[ctr]

		if char == '[' {
			offset := 1
			for regex[ctr+offset] != ']' {
				offset++
			}

			res.AddExp(parseClass(regex[ctr+1 : ctr+offset]))
			ctr += offset
		} else if char == '(' {
			offset, open := 0, 1
			for open > 0 {
				offset++

				if regex[ctr+offset] == '(' {
					open++
				} else if regex[ctr+offset] == ')' {
					open--
				}
			}

			res.AddExp(parseGroup(regex[ctr+1:ctr+offset], res.vars))
			ctr += offset
		} else if char == '?' {
			lastExp := res.PopExp()
			res.AddExp(NewExpOptional(lastExp))
		} else if char == '*' {
			if regex[ctr+1] != '{' {
				panic("Invalid count specifier given for repeating expression, expect *{4} or *{16} for instance!")
			} else {
				ctr++
			}

			offset := 2
			for regex[ctr+offset] != '}' {
				offset++
			}

			count, lastExp := parseInt(regex[ctr+1:ctr+offset]), res.PopExp()
			res.AddExp(NewRepExpZero(lastExp, count))
			ctr += offset
		} else if char == '+' {
			if regex[ctr+1] != '{' {
				panic("Invalid count specifier given for repeating expression, expect *{4} or *{16} for instance!")
			} else {
				ctr++
			}

			offset := 2
			for regex[ctr+offset] != '}' {
				offset++
			}

			count, lastExp := parseInt(regex[ctr+1:ctr+offset]), res.PopExp()
			res.AddExp(NewRepExpOne(lastExp, count))
			ctr += offset
		} else if char == '{' {
			offset := 2
			for regex[ctr+offset] != '}' {
				offset++
			}

			count, lastExp := parseInt(regex[ctr+1:ctr+offset]), res.PopExp()
			res.AddExp(NewRepExpFixed(lastExp, count))
			ctr += offset
		} else if char == '|' {
			lastExp := res.PopExp()

			if regex[ctr+1] == '(' {
				offset, open := 1, 1
				for open > 0 {
					offset++

					if regex[ctr+1+offset] == '(' {
						open++
					} else if regex[ctr+1+offset] == ')' {
						open--
					}
				}

				switch lastExp := lastExp.(type) {
				case OrExpGroup:
					lastExp.AddExp(parseGroup(regex[ctr+2:ctr+1+offset], res.vars))

					res.AddExp(lastExp)
					ctr += offset + 1
				default:
					orExp := NewOrExpGroup()
					orExp.AddExp(lastExp)
					orExp.AddExp(parseGroup(regex[ctr+2:ctr+1+offset], res.vars))

					res.AddExp(orExp)
					ctr += offset + 1
				}
			} else {
				res.AddExp(NewFixedExp(string(regex[ctr+1])))
				ctr++
			}
		} else if char == '\\' {
			res.AddExp(NewFixedExp(string(regex[ctr+1])))
			ctr++
		} else if char == '#' {
			varKey := regex[ctr+1]
			if _, ok := res.vars[string(varKey)]; ok {
				res.AddExp(NewExpVariable(string(varKey), res.vars, nil))
			} else {
				lastExp := res.PopExp()
				res.AddExp(NewExpVariable(string(varKey), res.vars, &lastExp))
				res.vars[string(varKey)] = "F"
			}

			ctr++
		} else {
			res.AddExp(NewFixedExp(string(char)))
		}

		ctr++
	}

	if res.Len() == 1 {
		return res.PopExp()
	}

	return res
}

func Compile(regex string) Exp {
	return parseGroup([]rune(regex), nil)
}
