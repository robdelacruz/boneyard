package search

import (
	"log"

	"github.com/blevesearch/bleve"
)

func BleveIndex(dir string) (bleve.Index, error) {
	idx, err := bleve.Open(dir)
	if err == bleve.ErrorIndexPathDoesNotExist {
		log.Printf("Creating new search index in %s...\n", dir)

		// Create new bleve index
		m := bleve.NewIndexMapping()
		idx, err = bleve.New(dir, m)
		if err != nil {
			log.Printf("Error creating search index in %s (%s)\n", dir, err)
			return nil, err
		}
	} else if err != nil {
		log.Printf("Error opening search index in %s (%s)\n", dir, err)
		return nil, err
	}

	return idx, nil
}
