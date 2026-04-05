package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/dansonserge/DaFileService/controllers"
)

func SetupRoutes(router *gin.Engine, fileCtrl *controllers.FileController) {
	api := router.Group("/api/v1")
	{
		files := api.Group("/files")
		{
			files.POST("/upload", fileCtrl.Upload)
			files.GET("/:bucket/*key", fileCtrl.Download) // Using wildcard for deep paths
			files.GET("/preview/:bucket/*key", fileCtrl.GetPreview)
			files.DELETE("/:bucket/*key", fileCtrl.Delete)
		}

		api.GET("/buckets", fileCtrl.ListBuckets)
	}
}
