package media

import "testing"

func TestIsCloudinaryDeliveryURL(t *testing.T) {
	t.Parallel()

	cloud := "demo"
	valid := "https://res.cloudinary.com/demo/image/upload/v123/boms/products/loaf.jpg"
	if !IsCloudinaryDeliveryURL(cloud, valid) {
		t.Fatalf("expected valid delivery URL")
	}

	cases := []string{
		"",
		"http://res.cloudinary.com/demo/image/upload/x.jpg",
		"https://evil.example/demo/image/upload/x.jpg",
		"https://res.cloudinary.com/other/image/upload/x.jpg",
		"javascript:alert(1)",
	}
	for _, raw := range cases {
		if IsCloudinaryDeliveryURL(cloud, raw) {
			t.Fatalf("expected invalid for %q", raw)
		}
	}
}

func TestIsCloudinaryDeliveryURLInFolder(t *testing.T) {
	t.Parallel()

	cloud := "demo"
	folder := "boms/products"
	valid := []string{
		"https://res.cloudinary.com/demo/image/upload/v123/boms/products/loaf.jpg",
		"https://res.cloudinary.com/demo/image/upload/f_auto,q_auto,w_640/v1/boms/products/loaf.jpg",
		"https://res.cloudinary.com/demo/image/upload/boms/products/loaf.jpg",
	}
	for _, raw := range valid {
		if !IsCloudinaryDeliveryURLInFolder(cloud, folder, raw) {
			t.Fatalf("expected valid folder URL for %q", raw)
		}
	}

	invalid := []string{
		"https://res.cloudinary.com/demo/image/upload/v1/boms/other/loaf.jpg",
		"https://res.cloudinary.com/demo/image/upload/v1/evil/boms/products/loaf.jpg",
		"https://res.cloudinary.com/other/image/upload/v1/boms/products/loaf.jpg",
	}
	for _, raw := range invalid {
		if IsCloudinaryDeliveryURLInFolder(cloud, folder, raw) {
			t.Fatalf("expected invalid folder URL for %q", raw)
		}
	}
}
