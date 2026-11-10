package authz

import (
	"fmt"

	"go.temporal.io/server/common/authorization"
	"go.temporal.io/server/common/config"
)

type ClaimMapper struct{}

// systemCNs get cluster admin: admintools/UI, and Temporal's own services
// (they present the internode cert when calling the frontend).
var systemCNs = map[string]struct{}{
	"temporal-system-admin": {},
	"temporal-internode":    {},
}

func NewClaimMapper(_ *config.Config) authorization.ClaimMapper {
	return &ClaimMapper{}
}

func (c *ClaimMapper) GetClaims(authInfo *authorization.AuthInfo) (*authorization.Claims, error) {
	if authInfo == nil {
		return nil, fmt.Errorf("missing auth info")
	}
	if authInfo.TLSSubject == nil || authInfo.TLSSubject.CommonName == "" {
		return nil, fmt.Errorf("mTLS client certificate required")
	}

	cn := authInfo.TLSSubject.CommonName
	claims := &authorization.Claims{
		Subject:    cn,
		Namespaces: map[string]authorization.Role{},
	}

	if _, ok := systemCNs[cn]; ok {
		claims.System = authorization.RoleAdmin
		return claims, nil
	}

	claims.Namespaces[cn] = authorization.RoleWriter | authorization.RoleWorker
	return claims, nil
}
