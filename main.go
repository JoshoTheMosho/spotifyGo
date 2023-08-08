package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var store = sessions.NewCookieStore([]byte("43jl#@zj^xclj392$DfLZp(O*#)CSadF#$@LH21*D9@#$Zxc")) // Replace with a strong secret key

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		// Handle error
		log.Fatal("Error loading .env file")
	}

	router := gin.Default()

	router.LoadHTMLGlob("templates/*.tmpl") // Load templates from the "templates" folder
	router.Static("/assets", "./assets")

	router.GET("", func(c *gin.Context) {
		// Render a template
		c.HTML(http.StatusOK, "main.tmpl", gin.H{
			"title": "Joshua Machado",
		})
	})

	router.GET("/spotifyGo", func(c *gin.Context) {
		// Determine the user's authentication status
		isAuthenticated := isAuthenticated(c)

		// Render a template with conditional content
		c.HTML(http.StatusOK, "spotifyGo.tmpl", gin.H{
			"title":           "spotifyGo",
			"IsAuthenticated": isAuthenticated,
		})
	})

	// Route to initiate the Spotify OAuth2 flow
	router.GET("/auth/login", handleLogin)

	// Route to handle the Spotify callback with authorization code
	router.GET("/auth/callback", handleCallback)

	router.Run(":8080")
}

func handleLogin(c *gin.Context) {
	// Redirect users to the Spotify authorization URL
	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	redirectURI := os.Getenv("SPOTIFY_REDIRECT_URI")
	scopes := "user-read-email user-library-read" // Add necessary scopes

	authURL := "https://accounts.spotify.com/authorize" +
		"?response_type=code" +
		"&client_id=" + clientID +
		"&scope=" + scopes +
		"&redirect_uri=" + redirectURI

	c.Redirect(http.StatusSeeOther, authURL)
}

func handleCallback(c *gin.Context) {
	// Handle the callback, exchange code for tokens, and manage sessions
	code := c.Query("code")
	if code == "" {
		// Handle error
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing authorization code"})
		return
	}

	// Exchange the code for access and refresh tokens
	tokens, err := exchangeCodeForTokens(c, code)
	if err != nil {
		// Handle error: Token exchange failed
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token exchange failed"})
		return
	}

	// Store tokens securely in the session
	session, _ := store.Get(c.Request, "spotifyGo-session")
	session.Values["access_token"] = tokens.AccessToken
	session.Values["refresh_token"] = tokens.RefreshToken
	session.Save(c.Request, c.Writer)
	log.Println("Successfully stored session")

	// Redirect to the main page
	c.Redirect(http.StatusSeeOther, "/spotifyGo") // Replace "/" with the actual URL of your main page
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

func exchangeCodeForTokens(c *gin.Context, code string) (*TokenResponse, error) {
	// Set up the HTTP client
	client := &http.Client{}
	spotifyRedirectURI := os.Getenv("SPOTIFY_REDIRECT_URI")
	spotifyClientID := os.Getenv("SPOTIFY_CLIENT_ID")
	spotifyClientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")

	// Construct the request payload
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", spotifyRedirectURI) // Replace with your actual redirect URI

	// Create a POST request
	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", strings.NewReader(data.Encode()))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return nil, err
	}

	// Set headers (client ID and client secret)
	req.SetBasicAuth(spotifyClientID, spotifyClientSecret) // Replace with your actual client ID and secret
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Perform the request
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
		}
	}(resp.Body)

	// Decode the response JSON
	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return nil, err
	}

	return &tokenResp, nil
}

func isAuthenticated(c *gin.Context) bool {
	// Return true if authenticated, false otherwise
	session, _ := store.Get(c.Request, "spotifyGo-session")

	// Check if the access token exists in the session
	accessToken := session.Values["access_token"]
	log.Println(accessToken)
	return accessToken != nil
}
