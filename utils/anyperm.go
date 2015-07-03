package utils

import (
	"math/rand"
	"time"
)

// PermString permutates a slice of strings
func PermString(input []string) (permutation []string) {
	l := len(input)
	permutation = make([]string, l)
	rand.Seed(time.Now().Unix())
	p := rand.Perm(l)
	for i, j := range p {
		permutation[i] = input[j]
	}
	return
}
