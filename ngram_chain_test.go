package markov

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/kr/pretty"
)

var dummyRandFunc = func(int) int { return 0 }

func Test_ProcessText(t *testing.T) {
	t.Parallel()

	var getValidReader = func() io.Reader { return strings.NewReader("a b c d e f") }

	var tests = []struct {
		name string
		body io.Reader
		n    uint

		wantMap map[string]*candidates
		wantErr error
	}{
		{
			name: "ok - process trigrams",
			body: getValidReader(),
			n:    3,
			wantMap: map[string]*candidates{
				"a b": &candidates{
					words: []wordFrequency{
						{word: "c", frequency: 1},
					},
					occurrences: 1,
				},
				"b c": &candidates{
					words: []wordFrequency{
						{word: "d", frequency: 1},
					},
					occurrences: 1,
				},
				"c d": &candidates{
					words: []wordFrequency{
						{word: "e", frequency: 1},
					},
					occurrences: 1,
				},
				"d e": &candidates{
					words: []wordFrequency{
						{word: "f", frequency: 1},
					},
					occurrences: 1,
				},
			},
			wantErr: nil,
		},
		{
			name: "ok - process 4-grams",
			body: getValidReader(),
			n:    4,
			wantMap: map[string]*candidates{
				"a b c": &candidates{
					words: []wordFrequency{
						{word: "d", frequency: 1},
					},
					occurrences: 1,
				},
				"b c d": &candidates{
					words: []wordFrequency{
						{word: "e", frequency: 1},
					},
					occurrences: 1,
				},
				"c d e": &candidates{
					words: []wordFrequency{
						{word: "f", frequency: 1},
					},
					occurrences: 1,
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		var tt = tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var NGramChain = NGramChain{
				store:    map[string]*candidates{},
				randFunc: dummyRandFunc,
				lock:     &sync.RWMutex{},
				n:        tt.n,
			}

			var err = NGramChain.ProcessText(tt.body)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("got %v, want %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(NGramChain.store, tt.wantMap) {
				t.Errorf("got %v, want %v", NGramChain.store, tt.wantMap)
			}
		})
	}
}

func TestNGramChain_GenerateRandomText(t *testing.T) {
	t.Parallel()

	var getMap = func() NGramChain {
		return NGramChain{
			store: map[string]*candidates{
				"It's a": &candidates{
					words: []wordFrequency{
						{word: "trap", frequency: 1},
						{word: "wonderful", frequency: 5},
					},
					occurrences: 6,
				},
				"I am": &candidates{
					words: []wordFrequency{
						{word: "batman", frequency: 4},
					},
					occurrences: 4,
				},
				"a wonderful": &candidates{
					words: []wordFrequency{
						{word: "world.", frequency: 2},
						{word: "planet", frequency: 3},
						{word: "day", frequency: 2},
					},
					occurrences: 7,
				},
				"wonderful planet": &candidates{
					words: []wordFrequency{
						{word: "we", frequency: 9},
					},
					occurrences: 9,
				},
				"planet we": &candidates{
					words: []wordFrequency{
						{word: "live", frequency: 3},
					},
					occurrences: 3,
				},
				"we live": &candidates{
					words: []wordFrequency{
						{word: "on", frequency: 5},
						{word: "tomorrow", frequency: 1},
					},
					occurrences: 6,
				},
			},
			seeds:    []string{"I am", "It's a"},
			randFunc: func(int) int { return 1 },
			lock:     &sync.RWMutex{},
		}
	}

	var emptyMap, _ = NewNGramChain(3)

	var tests = []struct {
		name       string
		NGramChain NGramChain

		wantText string
	}{
		{
			name:       "ok - empty map ",
			NGramChain: *emptyMap,
			wantText:   "",
		},
		{
			name: "ok - no seeds",
			NGramChain: NGramChain{
				store: map[string]*candidates{
					"i am": &candidates{
						words: []wordFrequency{
							{word: "batman", frequency: 4},
						},
						occurrences: 4,
					},
				},
				seeds:    []string{},
				randFunc: func(int) int { return 0 },
				lock:     &sync.RWMutex{},
			},
			wantText: "i am batman.",
		},
		{
			name:       "ok - two bigrams",
			NGramChain: getMap(),
			wantText:   "It's a wonderful world.",
		},
		{
			name: "ok - multiple bigrams",
			NGramChain: func() NGramChain {
				var m = getMap()
				m.seeds = []string{"I am", "Nope", "It's a"}
				m.randFunc = func(int) int { return 2 }
				return m
			}(),
			wantText: "It's a wonderful planet we live on.",
		},
		{
			name: "ok - one ngram adding .",
			NGramChain: func() NGramChain {
				var m = getMap()
				m.randFunc = func(int) int { return 0 }
				return m
			}(),
			wantText: "I am batman.",
		},
	}

	for _, tt := range tests {
		var tt = tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var text = tt.NGramChain.GenerateRandomText(100)

			if !reflect.DeepEqual(text, tt.wantText) {
				t.Errorf("got %v, want %v", text, tt.wantText)
			}
		})
	}
}

func TestNGramChain_Next(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		name  string
		input string

		wantCandidate string
	}{
		{
			name:          "ok",
			input:         "I am",
			wantCandidate: "batman",
		},
		{
			name:          "ok - key doesn't exist",
			input:         "You are",
			wantCandidate: "",
		},
	}

	for _, tt := range tests {
		var tt = tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var chain = getValidChain()

			var candidate = chain.GetCandidate(tt.input)
			if !reflect.DeepEqual(candidate, tt.wantCandidate) {
				t.Errorf("got %v, want %v", candidate, tt.wantCandidate)
			}
		})
	}
}

func TestNGramChain_CandidateProbability(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		name      string
		key       string
		candidate string

		wantProbability float32
		wantErr         error
	}{
		{
			name:            "ok",
			key:             "I am",
			candidate:       "batman",
			wantProbability: 1.0,
			wantErr:         nil,
		},
		{
			name:            "ok - key doesn't exist",
			key:             "You are",
			candidate:       "batman",
			wantProbability: 0.0,
			wantErr:         errors.New("prefix does not exist"),
		},
		{
			name:            "ok - candidate doesn't exist",
			key:             "I am",
			candidate:       "groot",
			wantProbability: 0.0,
			wantErr:         nil,
		},
	}

	for _, tt := range tests {
		var tt = tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var chain = getValidChain()

			var probability, err = chain.CandidateProbability(tt.key, tt.candidate)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("got %v, want %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(probability, tt.wantProbability) {
				t.Errorf("got %v, want %v", probability, tt.wantProbability)
			}
		})
	}
}

func TestNGramChain_processNgram(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		name  string
		n     uint
		ngram []string

		wantMap NGramChain
		wantErr error
	}{
		{
			name:  "ok - new ngram with seed",
			n:     3,
			ngram: []string{"It's", "a", "trap"},
			wantMap: NGramChain{
				store: map[string]*candidates{
					"It's a": &candidates{
						words: []wordFrequency{
							{word: "trap", frequency: 1},
						},
						occurrences: 1,
					},
					"I am": &candidates{
						words: []wordFrequency{
							{word: "batman", frequency: 4},
						},
						occurrences: 4,
					},
				},
				seeds: []string{"I am", "It's a"},
			},
			wantErr: nil,
		},
		{
			name:  "ok - new ngram no seed",
			n:     3,
			ngram: []string{"maybe", "another", "time"},
			wantMap: NGramChain{
				store: map[string]*candidates{
					"maybe another": &candidates{
						words: []wordFrequency{
							{word: "time", frequency: 1},
						},
						occurrences: 1,
					},
					"I am": &candidates{
						words: []wordFrequency{
							{word: "batman", frequency: 4},
						},
						occurrences: 4,
					},
				},
				seeds: []string{"I am"},
			},
			wantErr: nil,
		},
		{
			name:  "ok - existing ngram new candidate",
			n:     3,
			ngram: []string{"I", "am", "groot"},
			wantMap: NGramChain{
				store: map[string]*candidates{
					"I am": &candidates{
						words: []wordFrequency{
							{word: "batman", frequency: 4},
							{word: "groot", frequency: 1},
						},
						occurrences: 5,
					},
				},
				seeds: []string{"I am"},
			},
			wantErr: nil,
		},
		{
			name:  "ok - existing ngram existing candidate",
			n:     3,
			ngram: []string{"I", "am", "batman"},
			wantMap: NGramChain{
				store: map[string]*candidates{
					"I am": &candidates{
						words: []wordFrequency{
							{word: "batman", frequency: 5},
						},
						occurrences: 5,
					},
				},
				seeds: []string{"I am"},
			},
			wantErr: nil,
		},
		{
			name:    "error - invalid input",
			n:       3,
			ngram:   []string{"I", "am"},
			wantMap: getValidChain(),
			wantErr: errors.New("error processing ngram, expected input length 3, got 2"),
		},
	}

	for _, tt := range tests {
		var tt = tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var NGramChain = getValidChain()
			NGramChain.n = tt.n

			var err = NGramChain.processNgram(tt.ngram)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("got %v, want %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(NGramChain.store, tt.wantMap.store) {
				t.Errorf("got %v, want %v", NGramChain.store, tt.wantMap.store)
			}

			if !reflect.DeepEqual(NGramChain.seeds, tt.wantMap.seeds) {
				t.Errorf("got %v, want %v", NGramChain.seeds, tt.wantMap.seeds)
			}
		})
	}
}

func TestNGramChain_getRandomBigram(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		name       string
		NGramChain NGramChain

		wantBigram string
	}{
		{
			name:       "ok - with seeds",
			NGramChain: getValidChain(),
			wantBigram: "I am",
		},
		{
			name: "ok - without seeds",
			NGramChain: func() NGramChain {
				var m = getValidChain()
				m.seeds = []string{}
				return m
			}(),
			wantBigram: "I am",
		},
	}

	for _, tt := range tests {
		var tt = tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var ngram = tt.NGramChain.getRandomNGram()
			if !reflect.DeepEqual(ngram, tt.wantBigram) {
				t.Errorf("got %v, want %v", ngram, tt.wantBigram)
			}

		})
	}
}

func Test_NewNGramChain(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		name string
		n    uint

		wantMap *NGramChain
		wantErr error
	}{
		{
			name:    "ok",
			n:       3,
			wantErr: nil,
		},
		{
			name:    "error - invalid n",
			n:       1,
			wantErr: errors.New("error initialising NGramChain: n must be at least 2"),
		},
	}

	for _, tt := range tests {
		var tt = tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var _, err = NewNGramChain(tt.n)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("got %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestNGramChain_Concurrency(t *testing.T) {
	t.Parallel()

	var trigrams = [][]string{
		{"a", "a", "b"},
		{"a", "a", "c"},
		{"a", "a", "d"},
	}

	var NGramChain, _ = NewNGramChain(3)

	var wg = &sync.WaitGroup{}

	// spawn the writers
	for _, trigram := range trigrams {
		wg.Add(1)
		go func(tr []string) {
			for x := 0; x < 25; x++ {
				NGramChain.processNgram(tr)
			}
			wg.Done()
		}(trigram)
	}

	// spawn the readers
	for i := 0; i < 75; i++ {
		wg.Add(1)
		go func() {
			NGramChain.GenerateRandomText(100)
			wg.Done()
		}()
	}

	wg.Wait()

	var wantMap = map[string]*candidates{
		"a a": &candidates{
			words: []wordFrequency{
				{word: "b", frequency: 25},
				{word: "c", frequency: 25},
				{word: "d", frequency: 25},
			},
			occurrences: 75,
		},
	}

	var words = NGramChain.store["a a"].words
	sort.Slice(words, func(i, j int) bool {
		return words[i].word < words[j].word
	})

	if !reflect.DeepEqual(NGramChain.store, wantMap) {
		t.Errorf("got %v, want %v", pretty.Sprint(NGramChain.store), pretty.Sprint(wantMap))
	}
}

func getValidChain() NGramChain {
	return NGramChain{
		store: map[string]*candidates{
			"I am": &candidates{
				words: []wordFrequency{
					{word: "batman", frequency: 4},
				},
				occurrences: 4,
			},
		},
		seeds:    []string{"I am"},
		randFunc: dummyRandFunc,
		lock:     &sync.RWMutex{},
		n:        3,
	}
}

func Example() {
	// Create a chain for trigrams (3-grams)
	chain, _ := NewNGramChain(3)

	// Parse text to process
	text := strings.NewReader(`
	I am batman. 
	I am groot. 
	I am your father. `)

	chain.ProcessText(text)

	chain.GenerateRandomText(10)

	chain.GetCandidate("I am")

	probability, err := chain.CandidateProbability("I am", "batman.")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%0.2f\n", probability)
	// Output:
	// 0.33

}
