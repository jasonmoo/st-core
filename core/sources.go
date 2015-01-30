package core

func GetSources() map[string]SourceSpec {
	sources := []SourceSpec{
		StreamSource(),
		KeyValueSource(),
	}

	library := make(map[string]SourceSpec)

	for _, s := range sources {
		library[s.Name] = s
	}

	return library
}
