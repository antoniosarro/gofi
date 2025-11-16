package search

import (
	"strings"
	"unicode"

	"github.com/antoniosarro/gofi/internal/domain/entry"
)

// Index pre-processes entry data for faster searching
type Index struct {
	Entry          *entry.Entry
	SearchableText string
	NameNormalized string
	CommentTokens  []string
	CategoryTokens []string
}

// Indexer manages search indices for entries
type Indexer struct {
	indices map[string]*Index // key: entry.Path
}

// NewIndexer creates a new indexer
func NewIndexer() *Indexer {
	return &Indexer{
		indices: make(map[string]*Index),
	}
}

// Build creates indices for all entries
func (idx *Indexer) Build(entries []*entry.Entry) {
	for _, e := range entries {
		idx.Add(e)
	}
}

// Add creates an index for a single entry
func (idx *Indexer) Add(e *entry.Entry) {
	index := &Index{
		Entry:          e,
		NameNormalized: strings.ToLower(e.Name),
		CommentTokens:  tokenize(e.Comment),
		CategoryTokens: e.Categories,
	}

	// Build searchable text (name + generic name + first 10 words of comment)
	parts := []string{e.Name}
	if e.GenericName != "" {
		parts = append(parts, e.GenericName)
	}
	if e.Comment != "" {
		words := strings.Fields(e.Comment)
		if len(words) > 10 {
			words = words[:10]
		}
		parts = append(parts, strings.Join(words, " "))
	}
	index.SearchableText = strings.Join(parts, " ")

	idx.indices[e.Path] = index
}

// Get retrieves the index for an entry
func (idx *Indexer) Get(path string) *Index {
	return idx.indices[path]
}

// GetAll returns all indices
func (idx *Indexer) GetAll() []*Index {
	indices := make([]*Index, 0, len(idx.indices))
	for _, index := range idx.indices {
		indices = append(indices, index)
	}
	return indices
}

// Remove removes an index
func (idx *Indexer) Remove(path string) {
	delete(idx.indices, path)
}

// Clear removes all indices
func (idx *Indexer) Clear() {
	idx.indices = make(map[string]*Index)
}

// Count returns the number of indices
func (idx *Indexer) Count() int {
	return len(idx.indices)
}

// tokenize splits text into searchable tokens
func tokenize(text string) []string {
	text = strings.ToLower(text)
	tokens := make([]string, 0)

	var current strings.Builder
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current.WriteRune(r)
		} else {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		}
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}
