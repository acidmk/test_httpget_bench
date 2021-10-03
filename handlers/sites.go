package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"test-bench/api"
	"test-bench/benchmarks"
	"time"
)

func GetSites(c *gin.Context) {
	q, ok := c.GetQuery("search")

	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "search parameter is required",
		})
		return
	}

	res, err := api.GetYandexSearchResults(q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("failed to get Yandex search result: %s", err),
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	bench, err := benchmarks.NewSitesBench(res)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("failed to benchmark search results: %s", err),
		})
		return
	}
	defer bench.Close()

	sites, err := bench.Run(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("failed to benchmark search results: %s", err),
		})
		return
	}

	b, err := json.Marshal(sites)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("failed to marshal data to json: %s", err),
		})
		return
	}

	c.Data(http.StatusOK, "application/json", b)
}
