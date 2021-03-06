package main

import (
	"fmt"
	"sort"
)

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func cmp(word, key string) int {
	minLen := min(len(word), len(key))

	if word[:minLen] == key[:minLen] {
		return minLen
	}

	var matchLen int
	for i := 0; i < minLen; i++ {
		if word[i] != key[i] {
			break
		}
		matchLen++
	}
	return matchLen
}

type result struct {
	str    string
	length int
}

func main() {
	names := []string{"name1", "Yu", "name3", "Yury", "name3", "Yur", ""}
	searchKey := "Yury"

	var results []result

	for _, name := range names {
		if l := cmp(name, searchKey); l > 0 {
			results = append(results, result{name, l})
		}
	}
	sort.Slice(results,
		func(i, j int) bool { return results[i].length > results[j].length })
	fmt.Println(results)
	// [{Yury 4} {Yur 3} {Yu 2}]
}
