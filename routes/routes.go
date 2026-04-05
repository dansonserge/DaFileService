package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/dansonserge/DaFileService/controllers"
)

func SetupRoutes(router *gin.Engine, fileCtrl *controllers.FileController) {
	// Professional Deployment Namespace: /api/file/v1
	apiFile := router.Group("/api/file/v1")
	{
		files := apiFile.Group("/files")
		{
			files.POST("/upload", fileCtrl.Upload)
			files.GET("/:bucket/*key", fileCtrl.Download) // Deep path support for clinical silos
			files.GET("/preview/:bucket/*key", fileCtrl.GetPreview)
			files.DELETE("/:bucket/*key", fileCtrl.Delete)
		}

		apiFile.GET("/buckets", fileCtrl.ListBuckets)
	}
}
