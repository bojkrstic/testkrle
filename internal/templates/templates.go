package templates

import (
	"html/template"
)

// Load parses all templates and returns a *template.Template
func Load(pattern string) (*template.Template, error) {
	funcMap := template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"seq": func(start, end int) []int {
			if end < start {
				return []int{}
			}
			s := make([]int, 0, end-start+1)
			for i := start; i <= end; i++ {
				s = append(s, i)
			}
			return s
		},
	}
	t, err := template.New("tmpl").Funcs(funcMap).ParseGlob(pattern)
	if err != nil {
		return nil, err
	}
	return t, nil
}
