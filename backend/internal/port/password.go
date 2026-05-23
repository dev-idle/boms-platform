package port

// PasswordHasher hashes and verifies passwords without embedding cost policy in callers.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(encoded, password string) error
	NeedsRehash(encoded string) bool
}
