package authentication

import "context"

type anonymousService struct{}

func (anonymousService) Authenticate(
	context.Context,
	AuthenticationRequest,
) (AuthenticationResult, error) {
	return AuthenticationResult{
		Success: true,
		Principal: &Principal{
			ID:        "anonymous",
			Name:      "anonymous",
			Anonymous: true,
		},
	}, nil
}
