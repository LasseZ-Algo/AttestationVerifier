package controllers

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func UploadReport(c *gin.Context) {
	file, err := c.FormFile("report")
	if err != nil {
		c.String(http.StatusBadRequest, "Failed to upload file: %v", err)
		return
	}

	log.Printf("Received file: %s", file.Filename)

	c.String(http.StatusOK, "File %s uploaded successfully.", file.Filename)
}

func VerifyReport(c *gin.Context) {
	reportID := c.Param("report_id")

	c.String(http.StatusOK, "Report %s verified successfully.", reportID)
}
