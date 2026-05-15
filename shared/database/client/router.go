package client

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	api "github.com/Arc-Services/Arc/shared/api/util"
	"github.com/Arc-Services/Arc/shared/database"
	"github.com/Arc-Services/Arc/shared/middleware"
	"github.com/dgraph-io/badger/v4"
	"github.com/gin-gonic/gin"
)

func Setup(r *gin.Engine) {
	group := r.Group("/internal/private/api/v1/database")
	group.Use(middleware.DataBase())
	group.Use(signatureVerification())
	group.POST("/", Handler)
}

func signatureVerification() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientID := c.GetHeader("X-Arc-Client")
		signature := c.GetHeader(HeaderSignature)
		timestamp := c.GetHeader(HeaderTimestamp)
		nonce := c.GetHeader(HeaderNonce)

		if signature == "" || timestamp == "" || nonce == "" {
			c.JSON(401, api.ToJson(api.CreateError(c, "missing request signature", 401)))
			c.Abort()
			return
		}

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(400, api.ToJson(api.CreateError(c, "failed to read request body", 400)))
			c.Abort()
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(body))

		if err := VerifySignature(clientID, signature, timestamp, nonce, body); err != nil {
			c.JSON(401, api.ToJson(api.CreateError(c, "invalid request signature", 401)))
			c.Abort()
			return
		}

		c.Next()
	}
}

func Handler(c *gin.Context) {
	db := middleware.MustGetClientDB(c)

	var req struct {
		Action string      `json:"action"`
		Key    string      `json:"key"`
		Value  interface{} `json:"value"`
		Prefix string      `json:"prefix"`
		TTL    int64       `json:"ttl"`
		Limit  int         `json:"limit"`
		Ops    []struct {
			Action string      `json:"action"`
			Key    string      `json:"key"`
			Value  interface{} `json:"value"`
			TTL    int64       `json:"ttl"`
		} `json:"ops"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, api.ToJson(api.CreateError(c, err.Error(), 400)))
		return
	}

	switch req.Action {
	case "get":
		handleGet(c, db, req.Key)
	case "set":
		handleSet(c, db, req.Key, req.Value, req.TTL)
	case "delete":
		handleDelete(c, db, req.Key)
	case "keys":
		handleKeys(c, db, req.Prefix)
	case "scan":
		handleScan(c, db, req.Prefix, req.Limit)
	case "batch":
		handleBatch(c, db, req.Ops)
	case "stats":
		handleStats(c, db)
	default:
		c.JSON(400, api.ToJson(api.CreateError(c, "unknown action", 400)))
	}
}

func handleGet(c *gin.Context, db *database.DB, key string) {
	value, err := db.Get(key)
	if err != nil {
		if err == badger.ErrKeyNotFound {
			c.JSON(404, api.ToJson(api.CreateError(c, "key not found", 404)))
			return
		}
		c.JSON(500, api.ToJson(api.CreateError(c, err.Error(), 500)))
		return
	}

	var jsonData interface{}
	if err := json.Unmarshal(value, &jsonData); err == nil {
		c.JSON(200, gin.H{"value": jsonData})
		return
	}

	c.JSON(200, gin.H{"value": string(value)})
}

func handleSet(c *gin.Context, db *database.DB, key string, value interface{}, ttl int64) {
	if ttl > 0 {
		data, err := json.Marshal(value)
		if err != nil {
			c.JSON(500, api.ToJson(api.CreateError(c, err.Error(), 500)))
			return
		}
		if err := db.SetTTL(key, data, time.Duration(ttl)*time.Second); err != nil {
			c.JSON(500, api.ToJson(api.CreateError(c, err.Error(), 500)))
			return
		}
	} else {
		if err := db.SetJSON(key, value); err != nil {
			c.JSON(500, api.ToJson(api.CreateError(c, err.Error(), 500)))
			return
		}
	}

	c.JSON(200, gin.H{"success": true})
}

func handleDelete(c *gin.Context, db *database.DB, key string) {
	if err := db.Delete(key); err != nil {
		c.JSON(500, api.ToJson(api.CreateError(c, err.Error(), 500)))
		return
	}
	c.JSON(200, gin.H{"success": true})
}

func handleKeys(c *gin.Context, db *database.DB, prefix string) {
	keys := db.Keys(prefix)
	c.JSON(200, gin.H{"keys": keys})
}

func handleScan(c *gin.Context, db *database.DB, prefix string, limit int) {
	if limit > 1000 {
		limit = 1000
	}

	results := make(map[string]interface{})
	count := 0

	db.Scan(prefix, func(key string, value []byte) error {
		if count >= limit && limit > 0 {
			return nil
		}

		var jsonData interface{}
		if err := json.Unmarshal(value, &jsonData); err == nil {
			results[key] = jsonData
		} else {
			results[key] = string(value)
		}

		count++
		return nil
	})

	c.JSON(200, gin.H{"results": results})
}

func handleBatch(c *gin.Context, db *database.DB, ops []struct {
	Action string      `json:"action"`
	Key    string      `json:"key"`
	Value  interface{} `json:"value"`
	TTL    int64       `json:"ttl"`
}) {
	results := make([]gin.H, len(ops))

	for i, op := range ops {
		result := gin.H{"key": op.Key}

		switch op.Action {
		case "set":
			if op.TTL > 0 {
				data, err := json.Marshal(op.Value)
				if err != nil {
					result["success"] = false
					result["error"] = err.Error()
				} else if err := db.SetTTL(op.Key, data, time.Duration(op.TTL)*time.Second); err != nil {
					result["success"] = false
					result["error"] = err.Error()
				} else {
					result["success"] = true
				}
			} else {
				if err := db.SetJSON(op.Key, op.Value); err != nil {
					result["success"] = false
					result["error"] = err.Error()
				} else {
					result["success"] = true
				}
			}

		case "delete":
			if err := db.Delete(op.Key); err != nil {
				result["success"] = false
				result["error"] = err.Error()
			} else {
				result["success"] = true
			}

		case "get":
			value, err := db.Get(op.Key)
			if err != nil {
				result["success"] = false
				result["error"] = err.Error()
			} else {
				var jsonData interface{}
				if err := json.Unmarshal(value, &jsonData); err == nil {
					result["value"] = jsonData
				} else {
					result["value"] = string(value)
				}
				result["success"] = true
			}

		default:
			result["success"] = false
			result["error"] = "unknown action"
		}

		if result["success"] == nil {
			result["success"] = true
		}

		results[i] = result
	}

	c.JSON(200, gin.H{"results": results})
}

func handleStats(c *gin.Context, db *database.DB) {
	keys := db.Keys("")
	c.JSON(200, gin.H{"count": len(keys)})
}
