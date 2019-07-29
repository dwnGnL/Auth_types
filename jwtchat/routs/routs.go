package routs

import(
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	// "github.com/jinzhu/gorm"
	"github.com/dgrijalva/jwt-go"
	"time"
	"fmt"
	"jwtchat/models"
	"jwtchat/db"
	"log"
)
var jwtKey = []byte("my_secret_key")
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

var onlineUsers = make(map[*websocket.Conn]string)
var users = make(map[*websocket.Conn]bool)
var broadcast = make(chan models.Message)
var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func Auth(c *gin.Context)  {
	c.HTML(http.StatusOK,"index.html",gin.H{"title":"authorization"})
}
var login string
func CheckLog(c *gin.Context){
	var account []models.Account
	var check bool
	db := db.GetDB()
	login=c.PostForm("login")
	pass:=c.PostForm("password")
	db.Find(&account)
	for _,acc:=range account{
		if login==acc.Login && pass==acc.Pass{
			logs := models.Logs{User:acc.Login}
			db.Create(&logs)
			expirationTime := time.Now().Add(5 * time.Minute)
			// Create the JWT claims, which includes the username and expiry time
			claims := &Claims{
				Username: login,
				StandardClaims: jwt.StandardClaims{
					// In JWT, the expiry time is expressed as unix milliseconds
					ExpiresAt: expirationTime.Unix(),
				},
			}
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			// Create the JWT string
			tokenString, err := token.SignedString(jwtKey)
			if err != nil {
				// If there is an error in creating the JWT return an internal server error
				fmt.Print("Failed to set jwtKey")
				return
			}

			// Finally, we set the client cookie for "token" as the JWT we just generated
			// we also set an expiry time which is the same as the token itself
			http.SetCookie(c.Writer, &http.Cookie{
				Name:    "token",
				Value:   tokenString,
				Expires: expirationTime,
			})
			c.Redirect(303,"http://localhost:8080/chat")
			check=true
		}
	}
	if !check{
		c.Redirect(303,"http://localhost:8080/auth")
	}
}

func Chat(c *gin.Context){
	// db := db.GetDB()
	_, err := c.Request.Cookie("token")
	// var logs models.Logs
	// db.First(&logs)
	// if logs.User==""{
	// 	c.HTML(http.StatusOK,"index.html",gin.H{"title":"authorization"})
	// }else{
	// 	c.HTML(http.StatusOK,"chat.html",gin.H{"title":"dsdsa"})
	// }
	if err != nil {
		if err == http.ErrNoCookie {
			// If the cookie is not set, return an unauthorized status
			c.HTML(http.StatusOK,"index.html",gin.H{"title":"authorization"})
			return
		}
		// For any other type of error, return a bad request status
		fmt.Print("Failed to give token")
		return
	}

	c.HTML(http.StatusOK,"chat.html",gin.H{"title":"Chat"})
	
}

func Wshandler(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("token")
	// Get the JWT string from the cookie
	tknStr := c.Value

	// Initialize a new instance of `Claims`
	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !tkn.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	conn, err := wsupgrader.Upgrade(w, r, nil)
	database:=db.GetDB()
	if err != nil {
		fmt.Printf("Failed to set websocket upgrade: %+v \n", err)
		return
	}
	// var logs models.Logs
	// database.First(&logs)
	// database.Delete(&logs)
	defer conn.Close()

	var history []models.History
	onlineUsers[conn] = claims.Username
	users[conn] = true

	database.Find(&history)

	for _, row:= range history{
		historyMsg := models.Message {
			User: row.User,
			Message: row.Message,
			Date: row.Date,
		}
		conn.WriteJSON(historyMsg)
	}

	for {
		var msg models.Message
		// Read in a new message as JSON and map it to a Message object
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(users, conn)
			delete(onlineUsers,conn)
			break
		}
		if msg.Message == "connect" {
			msg.User = onlineUsers[conn]
			msg.Message = "test.conn"
			conn.WriteJSON(msg)
		}else {
			msg.User = onlineUsers[conn]
			// Send the newly received message to the broadcast channel
			broadcast <- msg
		}
	}
}

func HandleMessages() {
	database:=db.GetDB()
	now := time.Now().Format("02.01.2006 15:04:05")
	for {
		var history models.History
		// Grab the next message from the broadcast channel
		msg := <-broadcast
		if msg.Message != " is online" {
			history.User = msg.User
			history.Message = msg.Message
			history.Date = now

			database.Create(&history)
		}

		// Send it out to every user that is currently connected
		for user := range users {
			msg.Date = now
			err := user.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				user.Close()
				delete(users, user)
				delete(onlineUsers,user)
			}
		}
	}
}
