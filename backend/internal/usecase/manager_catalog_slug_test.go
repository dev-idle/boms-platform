package usecase

import "testing"

func TestResolveManagerCatalogSlug(t *testing.T) {
	t.Parallel()

	t.Run("create derives from name when slug empty", func(t *testing.T) {
		t.Parallel()
		slug, err := resolveManagerCatalogSlug("Matcha Drinks", "", true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if slug != "matcha-drinks" {
			t.Fatalf("got %q", slug)
		}
	})

	t.Run("create uses explicit slug when provided", func(t *testing.T) {
		t.Parallel()
		slug, err := resolveManagerCatalogSlug("Matcha Drinks", "  custom-slug  ", true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if slug != "custom-slug" {
			t.Fatalf("got %q", slug)
		}
	})

	t.Run("edit rejects empty slug", func(t *testing.T) {
		t.Parallel()
		if _, err := resolveManagerCatalogSlug("Matcha Drinks", "", false); err == nil {
			t.Fatal("expected validation error")
		}
	})

	t.Run("edit normalizes provided slug", func(t *testing.T) {
		t.Parallel()
		slug, err := resolveManagerCatalogSlug("Matcha Drinks", "Updated-Slug", false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if slug != "updated-slug" {
			t.Fatalf("got %q", slug)
		}
	})
}
