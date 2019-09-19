package markov

import (
	"reflect"
	"testing"
)

func TestCandidates_processCandidate(t *testing.T) {
	t.Parallel()

	var getValidCandidates = func() *candidates {
		return &candidates{
			words: []wordFrequency{
				{word: "potato", frequency: 1},
			},
			occurrences: 1,
		}
	}

	var tests = []struct {
		name       string
		candidates *candidates
		input      string

		wantCandidates *candidates
	}{
		{
			name:       "ok - add new candidate",
			candidates: getValidCandidates(),
			input:      "banana",
			wantCandidates: func() *candidates {
				var c = getValidCandidates()
				c.occurrences = 2
				c.words = []wordFrequency{
					{word: "potato", frequency: 1},
					{word: "banana", frequency: 1},
				}
				return c
			}(),
		},
		{
			name:       "ok - existing candidate",
			candidates: getValidCandidates(),
			input:      "potato",
			wantCandidates: func() *candidates {
				var c = getValidCandidates()
				c.occurrences = 2
				c.words = []wordFrequency{
					{word: "potato", frequency: 2},
				}
				return c
			}(),
		},
	}

	for _, tt := range tests {
		var tt = tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.candidates.processCandidate(tt.input)

			if !reflect.DeepEqual(tt.candidates, tt.wantCandidates) {
				t.Errorf("got %v, want %v", tt.candidates, tt.wantCandidates)
			}
		})
	}
}

func TestCandidates_selectCandidate(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		name       string
		candidates *candidates
		randFunc   func(int) int

		wantWord string
	}{
		{
			name:       "ok -  potato",
			candidates: getValidCandidates(),
			randFunc:   func(int) int { return 0 },
			wantWord:   "potato",
		},
		{
			name:       "ok - tomato",
			candidates: getValidCandidates(),
			randFunc:   func(int) int { return 6 },
			wantWord:   "tomato",
		},
		{
			name: "invalid candidates occurrences",
			candidates: func() *candidates {
				var c = getValidCandidates()
				c.occurrences = 20
				return c
			}(),
			randFunc: func(int) int { return 15 },
			wantWord: "",
		},
	}

	for _, tt := range tests {
		var tt = tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var candidates = getValidCandidates()
			var word = candidates.selectCandidate(tt.randFunc)

			if !reflect.DeepEqual(word, tt.wantWord) {
				t.Errorf("got %v, want %v", word, tt.wantWord)
			}
		})
	}
}

func TestCandidates_getCandidate(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		name  string
		input string

		wantWordFreq *wordFrequency
	}{
		{
			name:  "ok",
			input: "banana",
			wantWordFreq: &wordFrequency{
				word:      "banana",
				frequency: 4,
			},
		},
		{
			name:         "candidate not found",
			input:        "platano",
			wantWordFreq: nil,
		},
	}

	for _, tt := range tests {
		var tt = tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var c = getValidCandidates()

			var wf = c.getCandidate(tt.input)
			if !reflect.DeepEqual(wf, tt.wantWordFreq) {
				t.Errorf("got %v, want %v", wf, tt.wantWordFreq)
			}
		})
	}
}

func getValidCandidates() *candidates {
	return &candidates{
		words: []wordFrequency{
			{word: "potato", frequency: 1},
			{word: "banana", frequency: 4},
			{word: "tomato", frequency: 5},
		},
		occurrences: 10,
	}
}
