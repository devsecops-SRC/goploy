package route

import (
	"errors"
	"net/http"
	"os"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/zhenorzz/goploy/core"
)

// CheckToken return token is vaild. Besides user/login router
func CheckToken(w http.ResponseWriter, r *http.Request) error {
	if "/user/login" == r.URL.Path {
		return nil
	}
	unPraseToken, ok := r.Header["Authorization"]
	if !ok || len(unPraseToken) == 0 {
		response := core.Response{Code: 1, Message: "非法请求"}
		response.Json(w)
		return errors.New("非法请求")
	}
	token, err := jwt.Parse(unPraseToken[0], func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("SIGN_KEY")), nil
	})

	if err == nil && token.Valid {
	} else {
		response := core.Response{Code: 401, Message: err.Error()}
		response.Json(w)
		return err
	}
	return nil
}