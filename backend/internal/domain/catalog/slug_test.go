package catalog

import "testing"

func TestNormalizeSlug(t *testing.T) {
	t.Parallel()

	slug, err := NormalizeSlug("  Sourdough-Loaf  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if slug != "sourdough-loaf" {
		t.Fatalf("got %q", slug)
	}

	if _, err := NormalizeSlug("Bad Slug!"); err == nil {
		t.Fatal("expected invalid slug error")
	}
}
