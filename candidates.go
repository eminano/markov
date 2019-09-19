package markov

// candidates represents a list of words that have followed a given bigram with
// their respective frequencies. It also keeps track of the total number of
// bigram occurences.
type candidates struct {
	words       []wordFrequency
	occurrences int
}

// wordFrequency represents a word and its frequency.
type wordFrequency struct {
	word      string
	frequency int
}

func (c *candidates) processCandidate(candidate string) {
	// increase occurences counter for the bigram
	c.occurrences++

	for i, wf := range c.words {
		// if the candidate already exists, increase frequency and stop looking
		if candidate == wf.word {
			c.words[i].frequency++
			return
		}
	}

	// if candidate doesn't exist, add it
	c.words = append(c.words, wordFrequency{word: candidate, frequency: 1})
}

func (c *candidates) selectCandidate(randFunc func(int) int) string {
	// get a random number in the range of the bigram occurences
	var randomPos = randFunc(c.occurrences)

	var counter = 0

	// for each word increase the counter based on their frequency to weight the
	// probability of the different candidates
	for _, wordFreq := range c.words {
		counter += wordFreq.frequency
		if counter > randomPos {
			return wordFreq.word
		}
	}

	// this should only happen if the occurences are somehow not aligned with
	// the frequencies (sum(frequencies) != occurences)
	//
	// It could be handled as an error, but I kept it as an empty string for
	// simplicity
	return ""
}

func (c *candidates) getCandidate(word string) *wordFrequency {
	for _, candidate := range c.words {
		if candidate.word == word {
			return &candidate
		}
	}

	return nil
}
