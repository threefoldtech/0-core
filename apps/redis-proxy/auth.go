package main

import (
	"fmt"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
)

//In checks if s is in l
func In(s string, l []string) bool {
	for _, t := range l {
		if strings.EqualFold(s, t) {
			return true
		}
	}

	return false
}

//AuthMethod handles authorization of a user
func AuthMethod(organizations []string, key string) (func(string) bool, error) {
	var scopes []string
	for _, organization := range organizations {
		scopes = append(
			scopes,
			fmt.Sprintf("user:memberof:%s", organization),
		)
	}

	pub, err := jwt.ParseECPublicKeyFromPEM([]byte(key))
	if err != nil {
		return nil, err
	}

	return func(token string) bool {
		log.Debugf("checking token: %s", token)
		t, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
			m, ok := t.Method.(*jwt.SigningMethodECDSA)
			if !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", t.Header["alg"])
			}
			if t.Header["alg"] != m.Alg() {
				return nil, fmt.Errorf("Unexpected signing algorithm: %v", t.Header["alg"])
			}
			return pub, nil
		})

		if err != nil {
			log.Errorf("JWT parse error: %s", err)
			return false
		}

		if !t.Valid {
			return false
		}

		claims := t.Claims.(jwt.MapClaims)

		if err := claims.Valid(); err != nil {
			log.Errorf("itsyouonline calim validation error: %s", err)
			return false
		}

		if azp, ok := claims["azp"].(string); ok {
			if In(azp, organizations) {
				return true
			}
		}

		var claimedScopes []interface{}
		if value, ok := claims["scope"]; ok {
			if claimedScopes, ok = value.([]interface{}); !ok {
				return false
			}
		} else {
			return false
		}

		for _, s := range claimedScopes {
			switch s := s.(type) {
			case string:
				if In(s, scopes) {
					return true
				}
			}
		}

		return false
	}, nil
}
