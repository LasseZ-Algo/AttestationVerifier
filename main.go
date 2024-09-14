package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings" // Added import for string manipulation

	"github.com/gin-gonic/gin"
	"github.com/google/go-sev-guest/verify"
	_ "github.com/mattn/go-sqlite3"
)

// Initialize the global database variable
var db *sql.DB

// Initialize the database and create the reports table if it doesn't exist
func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./reports.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	if err := db.Ping(); err != nil {
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

	if _, err := statement.Exec(); err != nil {
		log.Fatalf("Failed to execute statement: %v", err)
	}
}

// hashReport computes the SHA256 hash of the report string
func hashReport(report string) string {
	h := sha256.New()
	h.Write([]byte(report))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// checkAndAddReport verifies the report and stores it in the database if not already present
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
		return false, false, err // Handle other database errors
	} else {
		exists = true
	}

	return verified == 1, exists, nil
}

// decodeBase64 decodes a base64 encoded string
func decodeBase64(s string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("base64 decoding error: %v", err)
	}
	return data, nil
}

// extractStringField safely extracts a string field from a map
func extractStringField(data map[string]interface{}, field string) (string, bool) {
	if val, exists := data[field]; exists {
		if str, ok := val.(string); ok {
			return str, true
		}
	}
	return "", false
}

// extractNestedMap safely extracts a nested map field from a parent map
func extractNestedMap(data map[string]interface{}, parent string) (map[string]interface{}, bool) {
	if parentVal, exists := data[parent]; exists {
		if parentMap, ok := parentVal.(map[string]interface{}); ok {
			return parentMap, true
		}
	}
	return nil, false
}

// extractProductField attempts to extract the "Product" field from the attestation
func extractProductField(attestation map[string]interface{}) interface{} {
	if product, exists := attestation["Product"]; exists {
		return product
	}
	return nil
}

func main() {
	initDB()
	defer db.Close()

	r := gin.Default()

	r.POST("/verify", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			log.Printf("No file uploaded: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
			return
		}

		f, err := file.Open()
		if err != nil {
			log.Printf("Failed to open file: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
			return
		}
		defer f.Close()

		// Read the entire file into a byte slice
		fileBytes, err := io.ReadAll(f)
		if err != nil {
			log.Printf("Failed to read file: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
			return
		}

		// Convert byte slice to string for preprocessing
		bodyString := string(fileBytes)

		// **Preprocessing Step: Correct the malformed "pcrs" key**
		// Replace 'pcrs:{' with '"pcrs":{'
		fixedBodyString := strings.ReplaceAll(bodyString, `pcrs:{`, `"pcrs":{`)

		// Optionally, log the correction for debugging purposes
		// log.Printf("Fixed JSON: %s", fixedBodyString)

		// Convert the fixed string back to byte slice
		fixedBodyBytes := []byte(fixedBodyString)

		var jsonReport map[string]interface{}
		if err := json.Unmarshal(fixedBodyBytes, &jsonReport); err != nil {
			log.Printf("Invalid JSON file after correction: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON file after correction", "details": err.Error()})
			return
		}

		// Extract required fields dynamically
		source, sourceExists := extractStringField(jsonReport, "Source")
		protocol, protocolExists := extractStringField(jsonReport, "Protocol")
		attestation, attestationExists := extractNestedMap(jsonReport, "Attestation")

		if !sourceExists || !protocolExists || !attestationExists {
			log.Println("Missing required fields: Source, Protocol, or Attestation")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields: Source, Protocol, or Attestation"})
			return
		}

		report, ok := attestation["Report"].(string)
		if !ok || report == "" {
			log.Println("Missing or invalid 'Report' field in Attestation")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid 'Report' field in Attestation"})
			return
		}

		verified, alreadyExists, err := checkAndAddReport(report)
		if err != nil {
			log.Printf("Database or verification error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "An error occurred during verification.",
				"error":   err.Error(),
			})
			return
		}

		// Extract optional fields
		product := extractProductField(attestation)
		dataField, dataExists := attestation["Data"]
		runtimeField, runtimeExists := attestation["Runtime"]
		eventLogField, eventLogExists := attestation["EventLog"]
		quoteField, quoteExists := attestation["Quote"]

		// You can extend this section to handle additional dynamic fields as needed

		// Prepare report details
		reportDetails := gin.H{
			"Source":   source,
			"Protocol": protocol,
			"Product":  product,
		}

		// Add optional fields if they exist
		if dataExists {
			reportDetails["Data"] = dataField
		}
		if runtimeExists {
			reportDetails["Runtime"] = runtimeField
		}
		if eventLogExists {
			reportDetails["EventLog"] = eventLogField
		}
		if quoteExists {
			reportDetails["Quote"] = quoteField
		}

		response := gin.H{
			"message":          "Report verified",
			"already_verified": alreadyExists,
			"report_details":   reportDetails,
		}

		if !verified {
			response["message"] = "Report verification failed"
		}

		c.JSON(http.StatusOK, response)
	})

	// Serve static files (e.g., CSS, JS)
	r.Static("/static", "./static")

	// Serve index.html as the default page
	r.GET("/", func(c *gin.Context) {
		c.File("static/index.html")
	})

	// Start the server on port 3001
	if err := r.Run(":3001"); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
