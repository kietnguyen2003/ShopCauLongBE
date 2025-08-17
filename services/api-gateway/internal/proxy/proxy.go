package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func ProxyRequest(serviceURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		targetURL := serviceURL + c.Request.URL.Path
		if c.Request.URL.RawQuery != "" {
			targetURL += "?" + c.Request.URL.RawQuery
		}

		var body []byte
		if c.Request.Body != nil {
			body, _ = io.ReadAll(c.Request.Body)
		}

		req, err := http.NewRequest(c.Request.Method, targetURL, bytes.NewBuffer(body))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
			return
		}

		for key, values := range c.Request.Header {
			if strings.ToLower(key) != "content-length" {
				for _, value := range values {
					req.Header.Add(key, value)
				}
			}
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "Service unavailable"})
			return
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
			return
		}

		for key, values := range resp.Header {
			for _, value := range values {
				c.Header(key, value)
			}
		}

		c.Status(resp.StatusCode)
		c.Writer.Write(respBody)
	}
}

func ProxyWithUserContext(serviceURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User context required"})
			return
		}

		targetURL := serviceURL + c.Request.URL.Path
		
		if c.Request.Method == "GET" && c.Request.URL.Path == "/api/orders" {
			targetURL += fmt.Sprintf("?user_id=%v", userID)
			if c.Request.URL.RawQuery != "" {
				targetURL += "&" + c.Request.URL.RawQuery
			}
		} else if c.Request.URL.RawQuery != "" {
			targetURL += "?" + c.Request.URL.RawQuery
		}

		var body []byte
		if c.Request.Body != nil {
			body, _ = io.ReadAll(c.Request.Body)
			
			if c.Request.Method == "POST" && strings.Contains(c.Request.URL.Path, "/orders") {
				var orderData map[string]interface{}
				if err := json.Unmarshal(body, &orderData); err == nil {
					orderData["user_id"] = userID
					body, _ = json.Marshal(orderData)
				}
			}
		}

		req, err := http.NewRequest(c.Request.Method, targetURL, bytes.NewBuffer(body))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
			return
		}

		for key, values := range c.Request.Header {
			if strings.ToLower(key) != "content-length" {
				for _, value := range values {
					req.Header.Add(key, value)
				}
			}
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "Service unavailable"})
			return
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
			return
		}

		for key, values := range resp.Header {
			for _, value := range values {
				c.Header(key, value)
			}
		}

		c.Status(resp.StatusCode)
		c.Writer.Write(respBody)
	}
}