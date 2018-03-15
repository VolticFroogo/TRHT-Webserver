package middleware

import (
	"net/http"
	"time"

	"github.com/VolticFroogo/The-Rabbit-Hole-Tearoom/helpers"
	"github.com/VolticFroogo/The-Rabbit-Hole-Tearoom/middleware/myJWT"
	"github.com/VolticFroogo/The-Rabbit-Hole-Tearoom/models"
)

// Admin handles authentication for admin pages.
func Admin(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	authTokenString, err := r.Cookie("authToken")
	if err != nil {
		helpers.ThrowErr(w, "Reading cookie error", err)
		return
	}

	refreshTokenString, err := r.Cookie("refreshToken")
	if err != nil {
		helpers.ThrowErr(w, "Reading cookie error", err)
		return
	}

	if authTokenString.Value != "" {
		authTokenValid, err := myJWT.CheckToken(authTokenString.Value, "", false, false)
		if err != nil {
			helpers.ThrowErr(w, "Checking token error", err)
			return
		}

		if authTokenValid {
			next(w, r)
			return
		}
	}

	if refreshTokenString.Value != "" {
		refreshTokenValid, err := myJWT.CheckToken(refreshTokenString.Value, "", true, false)
		if err != nil {
			helpers.ThrowErr(w, "Checking token error", err)
			return
		}

		if refreshTokenValid {
			newAuthTokenString, newRefreshTokenString, newCsrfSecret, err := myJWT.RefreshTokens(refreshTokenString.Value)
			if err != nil {
				helpers.ThrowErr(w, "Creating new tokens error", err)
				return
			}

			WriteNewAuth(w, r, newAuthTokenString, newRefreshTokenString, newCsrfSecret)

			next(w, r)
			return
		}
	}

	RedirectToHome(w, r)
}

// Form is the function used to protect forms.
func Form(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	authTokenString, err := r.Cookie("authToken")
	if err != nil {
		helpers.ThrowErr(w, "Reading cookie error", err)
		return
	}

	refreshTokenString, err := r.Cookie("refreshToken")
	if err != nil {
		helpers.ThrowErr(w, "Reading cookie error", err)
		return
	}

	csrfSecret := r.FormValue("csrfSecret")

	if authTokenString.Value != "" {
		authTokenValid, err := myJWT.CheckToken(authTokenString.Value, csrfSecret, false, true)
		if err != nil {
			helpers.ThrowErr(w, "Checking token error", err)
			return
		}

		if authTokenValid {
			next(w, r)
			return
		}
	}

	if refreshTokenString.Value != "" {
		refreshTokenValid, err := myJWT.CheckToken(refreshTokenString.Value, csrfSecret, true, true)
		if err != nil {
			helpers.ThrowErr(w, "Checking token error", err)
			return
		}

		if refreshTokenValid {
			newAuthTokenString, newRefreshTokenString, newCsrfSecret, err := myJWT.RefreshTokens(refreshTokenString.Value)
			if err != nil {
				helpers.ThrowErr(w, "Creating new tokens error", err)
				return
			}

			WriteNewAuth(w, r, newAuthTokenString, newRefreshTokenString, newCsrfSecret)

			next(w, r)
			return
		}
	}

	RedirectToHome(w, r)
}

// AJAX is the function used to protect AJAX requests.
func AJAX(w http.ResponseWriter, r *http.Request, data models.AJAXData) (valid bool) {
	valid = false

	authTokenString, err := r.Cookie("authToken")
	if err != nil {
		helpers.ThrowErr(w, "Reading cookie error", err)
		return
	}

	refreshTokenString, err := r.Cookie("refreshToken")
	if err != nil {
		helpers.ThrowErr(w, "Reading cookie error", err)
		return
	}

	if authTokenString.Value != "" {
		authTokenValid, err := myJWT.CheckToken(authTokenString.Value, data.CsrfSecret, false, true)
		if err != nil {
			helpers.ThrowErr(w, "Checking token error", err)
			return
		}

		if authTokenValid {
			return true
		}
	}

	if refreshTokenString.Value != "" {
		refreshTokenValid, err := myJWT.CheckToken(refreshTokenString.Value, data.CsrfSecret, true, true)
		if err != nil {
			helpers.ThrowErr(w, "Checking token error", err)
			return
		}

		if refreshTokenValid {
			newAuthTokenString, newRefreshTokenString, newCsrfSecret, err := myJWT.RefreshTokens(refreshTokenString.Value)
			if err != nil {
				helpers.ThrowErr(w, "Creating new tokens error", err)
				return
			}

			WriteNewAuth(w, r, newAuthTokenString, newRefreshTokenString, newCsrfSecret)

			return true
		}
	}

	return
}

// WriteNewAuth writes authentication to a user's browser.
func WriteNewAuth(w http.ResponseWriter, r *http.Request, authTokenString, refreshTokenString, csrfSecret string) {
	expiration := time.Now().Add(time.Hour * 24 * 365)

	cookie := http.Cookie{Name: "authToken", Value: authTokenString, Expires: expiration, Path: "/", HttpOnly: true, Secure: true}
	http.SetCookie(w, &cookie)

	cookie = http.Cookie{Name: "refreshToken", Value: refreshTokenString, Expires: expiration, Path: "/", HttpOnly: true, Secure: true}
	http.SetCookie(w, &cookie)

	cookie = http.Cookie{Name: "csrfSecret", Value: csrfSecret, Expires: expiration, Path: "/", HttpOnly: true, Secure: true}
	http.SetCookie(w, &cookie)

	return
}

// RedirectToHome redirects the client to home.
func RedirectToHome(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
