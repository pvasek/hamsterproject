package server

import (
	"io"
	"text/template"

	"github.com/labstack/echo/v4"
)

// Template for storing templates
type Template struct {
	templates *template.Template
}

// Render renders the template
func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
