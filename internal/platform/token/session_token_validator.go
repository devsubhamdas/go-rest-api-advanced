package token

type AccessTokenValidator struct {
	tokenManager *Manager
}

func NewAccessTokenValidator(tm *Manager) *AccessTokenValidator {
	return &AccessTokenValidator{
		tokenManager: tm,
	}
}

func (v *AccessTokenValidator) Validate(raw string) (string, error) {
	claims, err := v.tokenManager.Parse(raw, Access)
	if err != nil {
		return "", err
	}
	return claims.Subject, nil
}
