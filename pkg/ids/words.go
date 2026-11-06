package ids

import (
	"crypto/rand"
	_ "embed"
	"math/big"
	"strings"
)

// words: random words from the bundled list, "js-brisk-otter-lamp". Does
// not sort by time. Collisions are possible (about 270^3 ids for three
// words): the store retries Next inside the insert transaction.

//go:embed words.txt
var wordsTxt string

var wordList = strings.Fields(wordsTxt)

var wordSet = func() map[string]bool {
	m := map[string]bool{}
	for _, w := range wordList {
		m[w] = true
	}
	return m
}()

type wordsGen struct {
	n   int
	sep string
}

func (g wordsGen) Next(prefix, tag string) (string, error) {
	max := big.NewInt(int64(len(wordList)))
	parts := make([]string, g.n)
	for i := range parts {
		k, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		parts[i] = wordList[k.Int64()]
	}
	return join(prefix, tag, strings.Join(parts, g.sep)), nil
}

func (wordsGen) Sorted() bool { return false }

// Match accepts 2 to 8 listed words joined by one of the separators.
// Only listed words count, so a uuid whose groups happen to be all
// letters does not match.
func (wordsGen) Match(body string) bool {
	for _, sep := range []string{"-", "_", "."} {
		parts := strings.Split(body, sep)
		if len(parts) < 2 || len(parts) > 8 {
			continue
		}
		ok := true
		for _, p := range parts {
			if !wordSet[p] {
				ok = false
				break
			}
		}
		if ok {
			return true
		}
	}
	return false
}

// Retries reports whether the store should retry g's ids on a collision.
func Retries(g Generator) bool {
	_, ok := g.(wordsGen)
	return ok
}
