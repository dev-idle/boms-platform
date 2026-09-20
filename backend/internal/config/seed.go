package config

import (
	"fmt"
	"strings"
)

// DevAdminSeed is one development-only admin account created at API boot.
type DevAdminSeed struct {
	Email    string
	Password string
	FullName string
	Phone    string
}

// DevAdmins returns the primary seed admin followed by any extras (deduped by email).
func (c SeedConfig) DevAdmins() []DevAdminSeed {
	seen := make(map[string]struct{})
	var out []DevAdminSeed

	appendSeed := func(seed DevAdminSeed) {
		email := strings.TrimSpace(seed.Email)
		password := strings.TrimSpace(seed.Password)
		if email == "" || password == "" {
			return
		}
		normalized := strings.ToLower(email)
		if _, ok := seen[normalized]; ok {
			return
		}
		seen[normalized] = struct{}{}
		fullName := strings.TrimSpace(seed.FullName)
		if fullName == "" {
			fullName = "Development Admin"
		}
		out = append(out, DevAdminSeed{
			Email:    email,
			Password: password,
			FullName: fullName,
			Phone:    strings.TrimSpace(seed.Phone),
		})
	}

	appendSeed(DevAdminSeed{
		Email:    c.DevAdminEmail,
		Password: c.DevAdminPassword,
		FullName: c.DevAdminFullName,
		Phone:    c.DevAdminPhone,
	})
	for _, extra := range c.DevAdminExtras {
		appendSeed(extra)
	}

	return out
}

// parseDevAdminExtras parses SEED_DEV_ADMIN_EXTRAS entries:
// email|password|full_name[|phone] separated by semicolons.
func parseDevAdminExtras(raw string) ([]DevAdminSeed, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	var out []DevAdminSeed
	for _, entry := range strings.Split(raw, ";") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		parts := strings.Split(entry, "|")
		if len(parts) < 2 {
			return nil, fmt.Errorf("seed.dev_admin_extras entry must be email|password|full_name: %q", entry)
		}
		email := strings.TrimSpace(parts[0])
		password := strings.TrimSpace(parts[1])
		if email == "" {
			return nil, fmt.Errorf("seed.dev_admin_extras entry email is required: %q", entry)
		}
		if password == "" {
			return nil, fmt.Errorf("seed.dev_admin_extras entry password is required for %q", email)
		}
		fullName := "Development Admin"
		if len(parts) >= 3 {
			fullName = strings.TrimSpace(parts[2])
		}
		phone := ""
		if len(parts) >= 4 {
			phone = strings.TrimSpace(parts[3])
		}
		out = append(out, DevAdminSeed{
			Email:    email,
			Password: password,
			FullName: fullName,
			Phone:    phone,
		})
	}
	return out, nil
}
