package router

import (
	"lost-item/handlers"

	"github.com/gin-gonic/gin"
)

func Router() *gin.Engine {
	r := gin.Default()
	h := handlers.Handler{}
	h.Init()
	r.GET("/search", h.Search)
	r.GET("/item", h.ItemDetail)
	r.POST("/item", h.RegisterItem)
	r.DELETE("/item/:id", h.DeleteItem)
	r.POST("/parse", h.Parse)
	return r
}
