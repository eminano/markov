# markov
[![GoDoc](https://godoc.org/github.com/eminano/markov?status.svg)](https://godoc.org/github.com/eminano/markov)

Simple library using [markov chains](https://en.wikipedia.org/wiki/Markov_chain) for [n-gram](https://en.wikipedia.org/wiki/N-gram) modeling.

This library uses an in memory map storing the current state as key and the candidate states with the associated probability as value. 

Main features are: 
- Flexible ngram processing (support starting at 2-grams)
- Safe for concurrent use 
- Easy text processing support via io.Reader interface

## Usage

```go
package main

import (
	"fmt"
	"strings"

	"github.com/eminano/markov"
)

func main() {
	// Create a chain for trigrams (3-grams)
	chain, _ := markov.NewNGramChain(3)

	// Parse text to process
	text := strings.NewReader(`
	I am batman. 
	I am groot.
	I am your father`)

	chain.ProcessText(text)

	// generate random text based on input
	output := chain.GenerateRandomText(10)

	// get a random candidate for the prefix
	candidate := chain.GetCandidate("I am")

	// get the probability of a given candidate for a prefix
	probability := chain.CandidateProbability("I am", "batman.")
}
```

