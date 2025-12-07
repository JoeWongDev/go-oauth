package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var googleOauthConfig *oauth2.Config

func main() {
	// load the .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	googleOauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  "http://localhost:8080/callback",
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}

	// Basic endpoint
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/callback", handleCallback)
	http.HandleFunc("/logout", handleLogout)

	fmt.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	// read cookie
	cookie, err := r.Cookie("session_token")

	html := `<html><body><a href="/login">Google Login</a></body></html>`

	if err == nil {
		html = fmt.Sprintf(`<html><body>
		<h1>Welcome, %s!</h1>
		<a href="/logout">Logout</a>
		</body></html>`, cookie.Value)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, html)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	// state should be update as real random func, but set it as fixed string for testing
	url := googleOauthConfig.AuthCodeURL("random")
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	// check the state
	if r.FormValue("state") != "random" {
		fmt.Println("State is not valid")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// get the code from URL
	code := r.FormValue("code")

	// use the code to exchange token
	token, err := googleOauthConfig.Exchange(r.Context(), code)
	if err != nil {
		fmt.Printf("Could not get token: %s\n", err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// create client that use the token automatically
	client := googleOauthConfig.Client(r.Context(), token)

	// make the request for user info
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	// output user info
	data, _ := io.ReadAll(resp.Body)

	cookie := &http.Cookie{
		Name:  "session_token",
		Value: "User_Logged_In",
		Path:  "/",
	}
	http.SetCookie(w, cookie)

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)

	fmt.Fprintf(w, "UserInfo: %s\n", data)
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:   "session_token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
