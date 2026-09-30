package email

import (
	"bytes"
	"embed"
	"fmt"
	htmltemplate "html/template"
	texttemplate "text/template"

	"github.com/boms/backend/internal/port"
)

//go:embed templates
var templates embed.FS

// frame is what every email shows around its own part: the header, a heading
// and lead, one action and the footer.
type frame struct {
	Brand       brand
	Subject     string
	Preheader   string
	Eyebrow     string
	Heading     string
	Greeting    string
	Lead        []string
	ActionLabel string
	ActionURL   string
	FooterNote  string
}

// emailTemplates is one kind of email: its own "details" and "after" parts in
// the shared layout, as HTML and as plain text.
type emailTemplates struct {
	html *htmltemplate.Template
	text *texttemplate.Template
}

func parseEmail(name string) (emailTemplates, error) {
	html, err := htmltemplate.ParseFS(templates, "templates/layout.html", "templates/"+name+".html")
	if err != nil {
		return emailTemplates{}, fmt.Errorf("parse %s email html: %w", name, err)
	}
	text, err := texttemplate.ParseFS(templates, "templates/layout.txt", "templates/"+name+".txt")
	if err != nil {
		return emailTemplates{}, fmt.Errorf("parse %s email text: %w", name, err)
	}
	return emailTemplates{html: html, text: text}, nil
}

func (t emailTemplates) render(view any, to, toName, subject string) (port.Email, error) {
	var html, text bytes.Buffer
	if err := t.html.ExecuteTemplate(&html, "layout.html", view); err != nil {
		return port.Email{}, fmt.Errorf("render email html: %w", err)
	}
	if err := t.text.ExecuteTemplate(&text, "layout.txt", view); err != nil {
		return port.Email{}, fmt.Errorf("render email text: %w", err)
	}
	return port.Email{To: to, ToName: toName, Subject: subject, Text: text.String(), HTML: html.String()}, nil
}
