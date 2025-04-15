package jwtoken

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateJWToken(t *testing.T) {
	var teatCases = []struct {
		Name      string
		Claims    CustomClaims
		WantToken string
		WantErr   error
	}{
		{
			Name: "suc",
			Claims: CustomClaims{
				Name: "test",
			},
			WantToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoidGVzdCJ9.E4ib3Yz0kKRZ8BSgD5te_GRBRxpSp7MgwTOrD7KBZYY",
			WantErr:   nil,
		},
	}

	for _, tc := range teatCases {
		t.Run(tc.Name, func(t *testing.T) {
			token, err := NewJWToken().CreateJWToken(tc.Claims)
			assert.Equal(t, tc.WantErr, err)
			assert.Equal(t, tc.WantToken, token)
		})
	}
}

func TestParesJWToken(t *testing.T) {
	var teatCases = []struct {
		Name      string
		Token     string
		WantClaim CustomClaims
		WantErr   error
	}{
		{
			Name: "suc",
			WantClaim: CustomClaims{
				Name: "test",
			},
			Token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoidGVzdCJ9.E4ib3Yz0kKRZ8BSgD5te_GRBRxpSp7MgwTOrD7KBZYY",
			WantErr: nil,
		},
	}

	for _, tc := range teatCases {
		t.Run(tc.Name, func(t *testing.T) {
			claim, err := NewJWToken().ParesJWToken(tc.Token)
			assert.Equal(t, tc.WantErr, err)
			assert.Equal(t, tc.WantClaim.Name, claim.Name)
		})
	}
}
