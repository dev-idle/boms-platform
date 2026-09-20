package config

import "testing"

func TestSeedConfig_DevAdmins_PrimaryAndExtras(t *testing.T) {
	t.Parallel()

	cfg := SeedConfig{
		DevAdminEmail:    "admin@boms.com",
		DevAdminPassword: "ChangeMe@123",
		DevAdminFullName: "Development Admin",
		DevAdminExtras: []DevAdminSeed{
			{Email: "second.admin@example.com", Password: "ChangeMe@456", FullName: "Second Admin"},
		},
	}

	admins := cfg.DevAdmins()
	if len(admins) != 2 {
		t.Fatalf("expected 2 admins, got %d", len(admins))
	}
	if admins[0].Email != "admin@boms.com" {
		t.Fatalf("unexpected primary email: %q", admins[0].Email)
	}
	if admins[1].Email != "second.admin@example.com" {
		t.Fatalf("unexpected extra email: %q", admins[1].Email)
	}
}

func TestParseDevAdminExtras(t *testing.T) {
	t.Parallel()

	extras, err := parseDevAdminExtras("second.admin@example.com|ChangeMe@456|Second Admin")
	if err != nil {
		t.Fatal(err)
	}
	if len(extras) != 1 || extras[0].FullName != "Second Admin" {
		t.Fatalf("unexpected extras: %+v", extras)
	}

	_, err = parseDevAdminExtras("bad-entry")
	if err == nil {
		t.Fatal("expected parse error")
	}
}
