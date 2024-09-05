package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/go-sev-guest/verify"
	_ "github.com/mattn/go-sqlite3"
)

// JSONReport represents the structure of the incoming JSON report
// make dynamic type based on reportz??
type JSONReport struct {
	Version     int                    `json:"Version"`
	Source      string                 `json:"Source"`
	Protocol    string                 `json:"Protocol"`
	Instance    string                 `json:"Instance"`
	Attestation map[string]interface{} `json:"Attestation"`
}

var db *sql.DB

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./reports.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	statement, err := db.Prepare(`
    CREATE TABLE IF NOT EXISTS reports (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        report_hash TEXT UNIQUE,
        report TEXT,
        verified INTEGER
    )`)
	if err != nil {
		log.Fatalf("Failed to prepare statement: %v", err)
	}

	_, err = statement.Exec()
	if err != nil {
		log.Fatalf("Failed to execute statement: %v", err)
	}
}

func hashReport(report string) string {
	h := sha256.New()
	h.Write([]byte(report))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func checkAndAddReport(report string) (bool, bool, error) {
	reportHash := hashReport(report)

	var exists bool
	var verified int
	err := db.QueryRow("SELECT verified FROM reports WHERE report_hash=?", reportHash).Scan(&verified)
	if err == sql.ErrNoRows {
		exists = false
		rawReport, err := decodeBase64(report)
		if err != nil {
			log.Printf("Failed to decode base64: %v", err)
			return false, false, err
		}
		options := &verify.Options{}
		err = verify.RawSnpReport(rawReport, options)
		if err != nil {
			log.Printf("Verification failed: %v", err)
			verified = 0
		} else {
			verified = 1
		}

		_, err = db.Exec("INSERT INTO reports (report_hash, report, verified) VALUES (?, ?, ?)", reportHash, report, verified)
		if err != nil {
			log.Printf("Failed to insert into DB: %v", err)
			return false, false, err
		}
	} else if err != nil {
		return false, false, err // Handle error in querying the database
	} else {
		exists = true
	}

	return verified == 1, exists, nil
}

func decodeBase64(s string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("base64 decoding error: %v", err)
	}
	return data, nil
}

// TODO: add source, protocol and product to output
func main() {
	initDB()
	r := gin.Default()

	r.POST("/verify", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
			return
		}

		f, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
			return
		}
		defer f.Close()

		var jsonReport JSONReport
		if err := json.NewDecoder(f).Decode(&jsonReport); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON file"})
			return
		}

		report, ok := jsonReport.Attestation["Report"].(string)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Report field missing or invalid"})
			return
		}

		verified, alreadyExists, err := checkAndAddReport(report)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database or verification error", "details": err.Error()})
			return
		}

		response := gin.H{
			"message":          "Report verified",
			"already_verified": alreadyExists,
			"report_details": gin.H{
				"Source":   jsonReport.Source,
				"Protocol": jsonReport.Protocol,
				"Product":  jsonReport.Attestation["Product"],
			},
		}

		if !verified {
			response["message"] = "Report verification failed"
		}

		c.JSON(http.StatusOK, response)
	})

	// serve stylesheet and scripts
	r.Static("/static", "./static")

	// serve index.html as the default page:
	r.GET("/", func(c *gin.Context) {
		c.File("static/index.html")
	})

	if err := r.Run(":3001"); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
