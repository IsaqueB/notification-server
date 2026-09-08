package auth_test

import (
	"os"
	"testing"
	"time"

	"github.com/IsaqueB/notification-server/internal/auth"
	"github.com/IsaqueB/notification-server/internal/models"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	godotenv.Load("../../.env")

	exitCode := m.Run()
	os.Exit(exitCode)
}

func TestVerifyJWT_HS256(t *testing.T) {
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjbGllbnRfaWQiOiJ3ZXZlcnRvbiIsImV4cGlyZXNBdCI6IjIwMjYtMDktMDhUMTY6MDI6MzQuNzA5ODEtMDM6MDAifQ.tGHMJhDBE3NZnm4EEeTtpVoA08IvyED1WYMyu3M5GAs"
	clientId, err := auth.JWTAuthentication(token)
	require.Nil(t, err)
	require.Equal(t, "weverton", clientId)
}

func TestCreateJWT_HS256(t *testing.T) {
	clientId := "weverton"
	token, err := auth.CreateJWT_HS256(clientId, []models.Topic{}, time.Now().Add(10*time.Second))
	require.Nil(t, err)
	require.NotEmpty(t, token)
}
