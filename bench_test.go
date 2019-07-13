package main

import "testing"

func BenchmarkRandom(b *testing.B) {
	runIterations(b.N, map[string]AI{
		"Random": makeRandomMove,
	})
}

func BenchmarkGreedy(b *testing.B) {
	runIterations(b.N, map[string]AI{
		"Greedy": makeGreedyMove,
	})
}

func BenchmarkRare(b *testing.B) {
	runIterations(b.N, map[string]AI{
		"Rare": makeRareMove,
	})
}

func BenchmarkGreedyMean(b *testing.B) {
	runIterations(b.N, map[string]AI{
		"GreedyMean": makeGreedyExpectedValueMove,
	})
}
