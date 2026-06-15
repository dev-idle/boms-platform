package cloudinary

import "testing"

func TestSignUpload_isDeterministic(t *testing.T) {
	t.Parallel()

	params := map[string]string{
		"folder":    "boms/products",
		"timestamp": "1710000000",
	}
	first := SignUpload(params, "secret")
	second := SignUpload(params, "secret")
	if first != second || len(first) != 40 {
		t.Fatalf("unexpected signature %q", first)
	}
}

func TestUploadEndpoint(t *testing.T) {
	t.Parallel()

	got := UploadEndpoint("demo")
	want := "https://api.cloudinary.com/v1_1/demo/image/upload"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
