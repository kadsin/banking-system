package auth_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/kadsin/banking-system/internal/auth"
	"github.com/stretchr/testify/require"
)

func TestJWTService_GenerateAndParse(t *testing.T) {
	svc := auth.NewJWTService("unit-test-secret", 60)
	userID := uuid.MustParse("018f4fb6-6f8a-7a58-9ca7-f5a5c1d04111")

	token, err := svc.GenerateAccessToken(userID)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	parsedUserID, err := svc.ParseAccessToken(token)
	require.NoError(t, err)
	require.Equal(t, userID, parsedUserID)
}

func TestJWTService_ParseAccessToken_InvalidSecret(t *testing.T) {
	svcA := auth.NewJWTService("secret-a", 60)
	svcB := auth.NewJWTService("secret-b", 60)

	token, err := svcA.GenerateAccessToken(uuid.New())
	require.NoError(t, err)

	_, err = svcB.ParseAccessToken(token)
	require.Error(t, err)
}
