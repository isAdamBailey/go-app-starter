// Package mailer sends transactional emails, such as magic-link sign-in
// links.
package mailer

import "context"

// Mailer sends transactional emails.
type Mailer interface {
	// SendMagicLink emails a sign-in link to toEmail.
	SendMagicLink(ctx context.Context, toEmail, link string) error
}

const (
	appName           = "the app" // TODO: replace with your app's name
	magicLinkSubject  = "Your sign-in link"
	magicLinkTTLHuman = "15 minutes"
)

func magicLinkBody(link string) (text, html string) {
	text = "Click the link below to sign in to " + appName + ". This link expires in " + magicLinkTTLHuman + " and can only be used once.\n\n" + link
	html = `<p>Click the link below to sign in to ` + appName + `. This link expires in ` + magicLinkTTLHuman + ` and can only be used once.</p><p><a href="` + link + `">` + link + `</a></p>`
	return text, html
}
