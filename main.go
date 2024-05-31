package main

import (
	"AttestationVerifier/controllers"
	"AttestationVerifier/database"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	r := gin.Default()

	// Datenbankverbindung und Migration
	database.Connect()
	database.Migrate()

	// Statische Dateien bereitstellen
	r.Static("/frontEnd", "./frontEnd")

	// HTML-Template rendern
	r.LoadHTMLGlob("frontEnd/templates/*")

	// Routen einrichten
	r.POST("/report/upload", controllers.UploadReport)
	r.GET("/report/:report_id", controllers.VerifyReport)

	// Server starten
	err := r.Run(":8080")
	if err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}

/*	todo 	-verification einlesen und abspeichern
/	todo 	-anzeigen
/	todo	-frontend designen
*/
