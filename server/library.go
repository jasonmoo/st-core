package server

import (
	"net/http"

	"github.com/nytlabs/st-core/core"
)

// naming confusion between "name" and "type" ~_~
type LibraryEntry struct {
	Type   string          `json:"type"`
	Source core.SourceType `json:"source"`
	// type if we need that later
}

func (s *Server) BlockLibraryHandler(w http.ResponseWriter, r *http.Request) {
	s.Lock()
	defer s.Unlock()

	l := []LibraryEntry{}

	for _, v := range s.library {
		l = append(l, LibraryEntry{
			v.Name,
			v.Source,
		})
	}

	writeJSON(w, l, http.StatusOK)
}

func (s *Server) SourceLibraryHandler(w http.ResponseWriter, r *http.Request) {
	s.Lock()
	defer s.Unlock()

	l := []LibraryEntry{}

	for _, v := range s.sourceLibrary {
		l = append(l, LibraryEntry{
			v.Name,
			v.Type,
		})
	}

	writeJSON(w, l, http.StatusOK)
}
