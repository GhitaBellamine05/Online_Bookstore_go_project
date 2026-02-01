package auth

import (
    "errors"
    "time"
    "online-bookstore/models"
    "github.com/golang-jwt/jwt/v5"
    "net/http"
)

var jwtKey = []byte("my-super-secret-key-change-in-production") 

func GenerateToken(userID int, username string, role string) (string, error) {
    expirationTime := time.Now().Add(24 * time.Hour)
    claims := &models.Claims{
        UserID: userID,
        Role:   role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(expirationTime),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Subject:   username,
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtKey)
}

func ValidateToken(tokenString string) (*models.Claims, error) {
    claims := &models.Claims{}
    token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
        return jwtKey, nil
    })
    if err != nil {
        return nil, err
    }
    if !token.Valid {
        return nil, errors.New("invalid token")
    }
    return claims, nil
}
func ValidateTokenFromRequest(r *http.Request) (*models.Claims, error) {
    authHeader := r.Header.Get("Authorization")
    if authHeader == "" {
        return nil, errors.New("missing authorization header")
    }
    
    var tokenString string
    if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
        tokenString = authHeader[7:]
    } else {
        return nil, errors.New("invalid authorization header format")
    }
    
    return ValidateToken(tokenString)
}