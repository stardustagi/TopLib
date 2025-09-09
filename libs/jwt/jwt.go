package jwt

import (
	"fmt"

	jwt "github.com/dgrijalva/jwt-go"
)

func JWTDecrypt(tokenString, secret string) (jwt.MapClaims, bool) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		fmt.Println(err)
		return nil, false
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, true
	} else {
		fmt.Println(err)
		return nil, false
	}
}

func JWTEncrypt(id string, mytoken string, key string) string {
	mySigningKey := []byte(key)
	type MyCustomClaims struct {
		ID    string `json:"id"`
		Token string `json:"token"`
		jwt.StandardClaims
	}
	// Create the Claims
	claims := MyCustomClaims{
		id,
		mytoken,
		jwt.StandardClaims{
			ExpiresAt: 0,
			Issuer:    "admin",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtStr, _ := token.SignedString(mySigningKey)
	return jwtStr
}
