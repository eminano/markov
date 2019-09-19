package markov

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"sync"
)

// NGramChain is an in memory map store that processes ngrams and uses n-1grams as
// keys and a list of candidates as values. It's safe for concurrent use
type NGramChain struct {
	store map[string]*candidates
	n     uint

	seeds    []string
	randFunc func(n int) int
	lock     *sync.RWMutex
}

// ProcessText will parse the input and split it to process the ngrams as
// configured by the chain constructor
func (c *NGramChain) ProcessText(text io.Reader) error {
	var scanner = bufio.NewScanner(text)
	scanner.Split(bufio.ScanWords)

	var ngramCount uint
	var ngram = make([]string, c.n)

	// process the first ngram
	for scanner.Scan() {
		ngram[ngramCount] = scanner.Text()
		ngramCount++

		if ngramCount == c.n {
			if err := c.processNgram(ngram); err != nil {
				return err
			}
			break
		}
	}

	for scanner.Scan() {
		var nextNgram = make([]string, len(ngram))
		copy(nextNgram, ngram[1:])

		nextNgram[len(nextNgram)-1] = scanner.Text()

		if err := c.processNgram(nextNgram); err != nil {
			return err
		}

		ngram = nextNgram
	}

	return nil
}

// GenerateRandomText will generate a random text using the learnt ngrams
// keeping random selection of candidates weighted by frequency. It will
// generate a maximum of maxWords, less if the chain ends earlier.
func (c *NGramChain) GenerateRandomText(maxWords uint) string {
	c.lock.RLock()
	defer c.lock.RUnlock()

	// if the map is empty, no text to generate
	if len(c.store) == 0 {
		return ""
	}

	// start with a random seed
	var ngram = c.getRandomNGram()

	var strBuilder strings.Builder
	strBuilder.WriteString(ngram)

	for i := uint(0); i < maxWords; i++ {
		var candidates, exists = c.store[ngram]
		if !exists {
			// if the ngram doesn't exist, end the text generation
			break
		}

		var candidate = candidates.selectCandidate(c.randFunc)

		// add the candidate to the output Builder
		strBuilder.WriteByte(' ')
		strBuilder.WriteString(candidate)

		// generate new ngram with the selected candidate
		var ngramSplit = strings.Split(ngram, " ")
		var newNgram = make([]string, len(ngramSplit))
		copy(newNgram, ngramSplit[1:])

		newNgram[len(newNgram)-1] = candidate
		ngram = strings.Join(newNgram, " ")
	}

	// Add a dot at the end (if not present already)
	if !strings.HasSuffix(strBuilder.String(), ".") {
		strBuilder.WriteByte('.')
	}

	return strBuilder.String()
}

// GetCandidate will select and return a candidate for the given n-1gram prefix. It will return an empty
// string if the prefix doesn't exist
func (c *NGramChain) GetCandidate(prefix string) string {
	c.lock.RLock()
	defer c.lock.RUnlock()

	var candidates, exists = c.store[prefix]
	if !exists {
		return ""
	}

	return candidates.selectCandidate(c.randFunc)
}

// CandidateProbability will check what the probability of a given candidate is
// for a given n-1gram prefix. If the candidate does not exist, 0 is returned. If
// the prefix does not exist, an error is returned
func (c *NGramChain) CandidateProbability(prefix string, candidate string) (float32, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	var candidates, exists = c.store[prefix]
	if !exists {
		return 0.0, errors.New("prefix does not exist")
	}

	var wordFreq = candidates.getCandidate(candidate)
	if wordFreq == nil {
		return 0.0, nil
	}

	return float32(wordFreq.frequency) / float32(candidates.occurrences), nil
}

// processNgram will extract the ngram and candidate from the input and
// either add it to the map if it doesn't exist or increase frequency/add the
// new candidate
func (c *NGramChain) processNgram(input []string) error {
	// in order to process the ngram we need n on input
	if len(input) != int(c.n) {
		return fmt.Errorf("error processing ngram, expected input length %d, got %d", c.n, len(input))
	}

	// construct the key for the ngram map by concatenating the n-1 first words on
	// input. Use space as separator to preserve the original message and
	// differentiate between cases like "a bc" and "ab c"
	var ngram = strings.Join(input[:len(input)-1], " ")
	var candidate = input[len(input)-1]

	// lock the map to prevent racy reads while the writes are ongoing
	c.lock.Lock()
	defer c.lock.Unlock()

	// if the ngram already exists add the candidate to its list
	if candidates, exists := c.store[ngram]; exists {
		candidates.processCandidate(candidate)
		return nil
	}

	// otherwise add it to the map
	var candidates = &candidates{}
	candidates.processCandidate(candidate)

	c.store[ngram] = candidates

	// if the ngram starts with upper case, add it to the seed ngram list
	if ngram[0] >= 'A' && ngram[0] <= 'Z' {
		c.seeds = append(c.seeds, ngram)
	}

	return nil
}

// getRandomNGram returns a random ngram from the internal map. It will use
// the seeds if available
func (c *NGramChain) getRandomNGram() string {
	var ngram string

	// if there are seeds use them
	if len(c.seeds) > 0 {
		return c.seeds[c.randFunc(len(c.seeds))]
	}

	// otherwise pick a random ngram from the map
	var pos = c.randFunc(len(c.store))
	for k := range c.store {
		if pos == 0 {
			ngram = k
			break
		}
		pos--
	}

	return ngram
}

// NewNGramChain will initialise an ngram chain. The n on input will determine
// the length of the ngrams processed by the chain to produce the key (n-1gram)
// and candidates
func NewNGramChain(n uint) (*NGramChain, error) {
	if n <= 1 {
		return nil, errors.New("error initialising NGramChain: n must be at least 2")
	}

	return &NGramChain{
		store: make(map[string]*candidates),
		// having the randFunc as a field of the NGramChain allows for testing with deterministic output
		randFunc: rand.Intn,
		lock:     &sync.RWMutex{},
		n:        n,
	}, nil
}
