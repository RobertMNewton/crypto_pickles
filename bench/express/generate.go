package express

import "github.com/emirpasic/gods/sets/hashset"

func Generate(exp Exp, count int) []string {
	res := make([]string, count)
	for i := 0; i < count; i++ {
		res[i] = exp.Get()
	}

	return res
}

func GenerateUnique(exp Exp, attempts int) []string {
	set := hashset.New()
	for i := 0; i < attempts; i++ {
		set.Add(exp.Get())
	}

	res := make([]string, set.Size())
	for i, val := range set.Values() {
		res[i] = val.(string)
	}

	return res
}
