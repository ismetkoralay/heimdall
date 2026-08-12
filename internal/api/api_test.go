package api

import (
	"testing"

	"github.com/ismetkoralay/heimdall/internal/provider"
	"github.com/stretchr/testify/assert"
)

func TestMapProviderRole(t *testing.T) {
	tests := []struct {
		name         string
		givenRole    ChatRequestRole
		expectedRole provider.Role
	}{
		{
			name:         "maps assistant role correctly",
			givenRole:    ChatRequestRoleAssistant,
			expectedRole: provider.RoleAssistant,
		},
		{
			name:         "maps user role correctly",
			givenRole:    ChatRequestRoleUser,
			expectedRole: provider.RoleUser,
		},
		{
			name:         "maps developer role correctly",
			givenRole:    ChatRequestRoleDeveloper,
			expectedRole: provider.RoleDeveloper,
		},
		{
			name:         "maps system role correctly",
			givenRole:    ChatRequestRoleSystem,
			expectedRole: provider.RoleSystem,
		},
		{
			name:         "falls back to unknown role for unknown roles",
			givenRole:    ChatRequestRole("dummy"),
			expectedRole: provider.RoleUnkown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			res := mapProviderRole(tt.givenRole)

			// Assert
			assert.Equal(t, tt.expectedRole, res)
		})
	}
}
