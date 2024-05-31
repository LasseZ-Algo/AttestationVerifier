package main

import (
	"AttestationVerifier/controllers"
	"AttestationVerifier/database"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func main() {
	r := gin.Default()

	r.Static("/frontEnd", "./frontEnd")
	r.LoadHTMLGlob("frontEnd/htmls/*")

	database.Connect()
	database.Migrate()

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	r.POST("/report/upload", controllers.UploadReport)
	r.GET("/report/:report_id", controllers.VerifyReport)

	err := r.Run(":8080")
	if err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}

/*	todo 	-text hochladen können
/	todo 	-verification einlesen und abspeichern
/	todo 	-anzeigen
/	todo	-frontend designen (soll ansehnlich sein)
*/
