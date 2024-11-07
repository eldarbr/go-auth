package encrypt

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"os"

	"github.com/eldarbr/go-auth/internal/service/myerrors"
	"github.com/golang-jwt/jwt"
)

var (
	ErrParsingToken = errors.New("couldn't parse the token")
	ErrWrongClaims  = errors.New("unknown claims type, cannot proceed")
)

type JWTService struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

type ClaimUserGroupRole struct {
	ServiceName string `json:"serviceName"`
	UserRole    string `json:"userRole"`
}

type AuthCustomClaims struct {
	Username string               `json:"username"`
	Groups   []ClaimUserGroupRole `json:"groups"`
}

type myCompletelaims struct {
	jwt.StandardClaims
	AuthCustomClaims
}

func NewJWTService(privatePath, publicPath string) (*JWTService, error) {
	privateKeyBytes, err := os.ReadFile(privatePath)
	if err != nil {
		return nil, fmt.Errorf("NewJWTService private key read failed: %w", err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("NewJWTService private key parse failed: %w", err)
	}

	publicKeyBytes, err := os.ReadFile(publicPath)
	if err != nil {
		return nil, fmt.Errorf("NewJWTService public key read failed: %w", err)
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("NewJWTService public key parse failed: %w", err)
	}

	return &JWTService{
		privateKey: privateKey,
		publicKey:  publicKey,
	}, nil
}

func (jwtService *JWTService) IssueToken(claims AuthCustomClaims) (string, error) {
	if jwtService == nil {
		return "", myerrors.ErrServiceNullPtr
	}

	completeClaims := myCompletelaims{
		AuthCustomClaims: claims,
		StandardClaims: jwt.StandardClaims{ //nolint:exhaustruct // other fields are not used.
			IssuedAt: jwt.TimeFunc().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS512, completeClaims)

	signedToken, err := token.SignedString(jwtService.privateKey)
	if err != nil {
		return "", fmt.Errorf("jwtService.IssueToken signing failed: %w", err)
	}

	return signedToken, nil
}

func (jwtService *JWTService) ValidateToken(tokenString string) (*AuthCustomClaims, error) {
	if jwtService == nil {
		return nil, myerrors.ErrServiceNullPtr
	}

	var (
		claims        *myCompletelaims
		tokenClaimsOk bool
	)

	//nolint:exhaustruct // Only the type is what matters.
	token, err := jwt.ParseWithClaims(tokenString, &myCompletelaims{}, func(_ *jwt.Token) (interface{}, error) {
		return jwtService.publicKey, nil
	})
	if err != nil {
		return nil, ErrParsingToken
	} else if claims, tokenClaimsOk = token.Claims.(*myCompletelaims); !tokenClaimsOk {
		return nil, ErrWrongClaims
	}

	return &claims.AuthCustomClaims, nil
}

func (claims AuthCustomClaims) Contain(requested ClaimUserGroupRole) bool {
	for _, claim := range claims.Groups {
		if (claim.ServiceName == requested.ServiceName || requested.ServiceName == "") &&
			(claim.UserRole == requested.UserRole || requested.UserRole == "") {
			return true
		}
	}

	return false
}
