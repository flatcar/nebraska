package admin

import (
	"crypto/md5" //nolint:gosec // retained only to verify pre-existing legacy secrets, see VerifyUserPassword
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"golang.org/x/crypto/bcrypt"

	"github.com/flatcar/nebraska/backend/pkg/api/internal/types"
)

const (
	// Realm used for basic authentication.
	Realm = "nebraska"

	// bcryptSecretPrefix is the prefix of every hash produced by
	// golang.org/x/crypto/bcrypt (e.g. "$2a$", "$2b$", "$2y$"). Secrets
	// stored before the move to bcrypt are exactly a 32-character
	// hex-encoded MD5 digest, which can never start with "$", so a
	// stored value's own shape tells the two schemes apart and no
	// separate scheme-marker column is needed.
	bcryptSecretPrefix = "$2"
)

// AddUser registers a user.
func (s *Service) AddUser(user *types.User) (*types.User, error) {
	query, _, err := goqu.Insert("users").
		Cols("username", "team_id", "secret").
		Vals(goqu.Vals{user.Username, user.TeamID, user.Secret}).
		Returning(goqu.T("users").All()).
		ToSQL()
	if err != nil {
		return nil, err
	}
	err = s.db.QueryRowx(query).StructScan(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// UpdateUserPassword updates the password of the provided user, storing
// it hashed with bcrypt.
func (s *Service) UpdateUserPassword(username, newPassword string) error {
	secret, err := s.GenerateUserSecret(username, newPassword)
	if err != nil {
		return err
	}
	query, _, err := goqu.Update("users").
		Set(goqu.Record{"secret": secret}).
		Where(goqu.C("username").Eq(username)).
		ToSQL()
	if err != nil {
		return err
	}
	result, err := s.db.Exec(query)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return types.ErrUpdatingPassword
	}

	return nil
}

// GenerateUserSecret hashes the password provided with bcrypt. The
// username parameter is kept only for backward compatibility with
// callers written for the previous md5(username:realm:password) scheme;
// it is not mixed into the bcrypt input, which already gets its own
// random salt and would otherwise silently count against bcrypt's
// 72-byte input limit, capping real passwords well below that.
func (s *Service) GenerateUserSecret(_, password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyUserPassword checks a username/password pair against a stored
// secret, accepting both formats a stored secret can be in:
//   - a bcrypt hash of the password, the current format;
//   - a legacy md5 hash of "username:realm:password", used before the
//     move to bcrypt.
//
// isLegacy reports whether the stored secret was in the legacy format,
// so that on a successful login the caller can re-hash the password
// with GenerateUserSecret and persist it via UpdateUserPassword,
// upgrading the account to bcrypt without requiring a password change.
func VerifyUserPassword(username, password, storedSecret string) (ok, isLegacy bool, err error) {
	if strings.HasPrefix(storedSecret, bcryptSecretPrefix) {
		err = bcrypt.CompareHashAndPassword([]byte(storedSecret), []byte(password))
		if err != nil {
			if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
				return false, false, nil
			}
			return false, false, err
		}
		return true, false, nil
	}

	legacySecret, err := generateLegacyMD5Secret(username, password)
	if err != nil {
		return false, false, err
	}
	match := subtle.ConstantTimeCompare([]byte(storedSecret), []byte(legacySecret)) == 1
	// isLegacy reflects the stored secret's format, not whether this
	// particular password matched it.
	return match, true, nil
}

// generateLegacyMD5Secret reproduces the pre-bcrypt secret format
// (username:realm:password), kept only so VerifyUserPassword can still
// check accounts that have not logged in since the move to bcrypt.
func generateLegacyMD5Secret(username, password string) (string, error) {
	h := md5.New() //nolint:gosec // legacy verification only, see VerifyUserPassword
	if _, err := io.WriteString(h, username+":"+Realm+":"+password); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// Authenticate verifies a username/password pair against the user's
// stored secret and, on success, transparently upgrades a legacy md5
// secret to bcrypt so the password is never checked against the
// weaker scheme again.
func (s *Service) Authenticate(username, password string) (*types.User, error) {
	user, err := s.GetUser(username)
	if err != nil {
		return nil, err
	}

	ok, isLegacy, err := VerifyUserPassword(username, password, user.Secret)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, types.ErrInvalidCredentials
	}

	if isLegacy {
		if err := s.UpdateUserPassword(username, password); err != nil {
			// The login itself already succeeded; failing to upgrade
			// the stored secret shouldn't fail it.
			l.Warn().Err(err).Str("username", username).Msg("failed to upgrade legacy password secret")
		}
	}

	return user, nil
}
