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

func TestSlugFromName(t *testing.T) {
	t.Parallel()

	slug, err := SlugFromName("Matcha Drinks")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if slug != "matcha-drinks" {
		t.Fatalf("got %q", slug)
	}

	slug, err = SlugFromName("Café")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if slug != "cafe" {
		t.Fatalf("got %q", slug)
	}

	if _, err := SlugFromName("!!!"); err == nil {
		t.Fatal("expected invalid slug error for empty derivation")
	}
}
