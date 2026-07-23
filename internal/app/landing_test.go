package app

import (
	"strings"
	"testing"
)

func TestLandingPageHTMLUsesAndEscapesFrontendURL(t *testing.T) {
	page := landingPageHTML(`https://example.com/?from=lumiere&name="guest"`)

	if !strings.Contains(page, `href="https://example.com/?from=lumiere&amp;name=%22guest%22"`) {
		t.Error("landing page does not contain the escaped frontend URL")
	}
	if !strings.Contains(page, "Music that feels") {
		t.Error("landing page does not contain the welcome heading")
	}
}
