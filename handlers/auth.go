package handlers

import (
	"crypto/sha256"
	"net/http"
	"os"
	"time"

	"github.com/crud-recipe/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
)

type AuthHandler struct {
	collection *mongo.Collection
	ctx        context.Context
}

func NewAuthHandler(ctx context.Context, collection *mongo.Collection) *AuthHandler {
	return &AuthHandler{
		collection: collection,
		ctx:        ctx,
	}
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

type JWTOutput struct {
	Token   string    `json:"token"`
	Expires time.Time `json:"expires"`
}

func (handler *AuthHandler) SignInHandler(c *gin.Context) {

	var user models.User

	if err := c.ShouldBindJSON(&user); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return

	}

	h := sha256.New()

	cur := handler.collection.FindOne(handler.ctx, bson.M{

		"username": user.Username,

		"password": string(h.Sum([]byte(user.Password))),
	})

	if cur.Err() != nil {

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})

		return

	}

	sessionToken := xid.New().String()

	session := sessions.Default(c)

	session.Set("username", user.Username)

	session.Set("token", sessionToken)

	session.Save()

	c.JSON(http.StatusOK, gin.H{"message": "User signed in"})

}

func (handler *AuthHandler) AuthMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {

		session := sessions.Default(c)

		sessionToken := session.Get("token")

		if sessionToken == nil {

			c.JSON(http.StatusForbidden, gin.H{

				"message": "Not logged",
			})

			c.Abort()

		}

		c.Next()

	}

}

func (handler *AuthHandler) RefreshHandler(c *gin.Context) {
	tokenValue := c.GetHeader("Authorization")
	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(tokenValue, claims,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	if tkn == nil || !tkn.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}
	if time.Unix(claims.ExpiresAt, 0).Sub(time.Now()) >
		30*time.Second {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token is not expired yet"})
		return
	}

	expirationTime := time.Now().Add(5 * time.Minute)
	claims.ExpiresAt = expirationTime.Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(os.Getenv("JWT_SECRET"))
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": err.Error()})
		return
	}
	jwtOutput := JWTOutput{
		Token:   tokenString,
		Expires: expirationTime,
	}
	c.JSON(http.StatusOK, jwtOutput)

}

func (handler *AuthHandler) SignOutHandler(c *gin.Context) {

	session := sessions.Default(c)

	session.Clear()

	session.Save()

	c.JSON(http.StatusOK, gin.H{"message": "Signed out..."})

}
