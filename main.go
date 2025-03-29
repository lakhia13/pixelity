package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	//"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"

	"nas-app/album"
	"nas-app/auth"
	"nas-app/config"
	"nas-app/media"
	"nas-app/search"
)

var (
	templates     *template.Template
	authService   *auth.AuthService
	mediaService  *media.MediaService
	albumService  *album.AlbumService
	searchService *search.SearchService
)

func init() {
	// Initialize database
	config.InitDB()

	// Initialize services
	authService = auth.NewAuthService(config.DB)
	mediaService = media.NewMediaService(config.DB)
	albumService = album.NewAlbumService(config.DB)
	searchService = search.NewSearchService(config.DB)

	// Parse templates
	templates = template.Must(template.ParseGlob("templates/*.html"))
}

func main() {
	router := gin.Default()

	// Serve static files
	router.Static("/static", "./static")
	router.Static("/uploads", "./uploads")
	router.Static("/thumbnails", "./thumbnails")

	// Auth routes
	router.GET("/login", showLoginPage)
	router.POST("/login", handleLogin)
	router.GET("/register", showRegisterPage)
	router.POST("/register", handleRegister)
	router.GET("/logout", handleLogout)

	// Protected routes
	authorized := router.Group("/")
	authorized.Use(authMiddleware())
	{
		authorized.GET("/", handleHome)
		//authorized.GET("/gallery", handleGallery)
		authorized.GET("/albums", handleAlbums)

		// Media routes
		authorized.POST("/upload", handleUpload)
		//authorized.GET("/media/:id", handleGetMedia)
		//authorized.DELETE("/media/:id", handleDeleteMedia)

		// Album routes
		authorized.POST("/albums", handleCreateAlbum)
		authorized.GET("/albums/:id", handleGetAlbum)
		authorized.PUT("/albums/:id", handleUpdateAlbum)
		authorized.DELETE("/albums/:id", handleDeleteAlbum)
		authorized.POST("/albums/:id/media", handleAddMediaToAlbum)
		authorized.DELETE("/albums/:id/media/:mediaId", handleRemoveMediaFromAlbum)

		// Search routes
		authorized.GET("/search", handleSearch)
		authorized.GET("/tags", handleGetTags)
	}

	fmt.Println("Server starting on :8080...")
	log.Fatal(router.Run(":8080"))
}

// Middleware
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("auth_token")
		if err != nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		userID, err := authService.ValidateToken(token)
		if err != nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}

func showLoginPage(c *gin.Context) {
	renderTemplate(c, "login.html", gin.H{
		"Title": "Login",
	})
}

func showRegisterPage(c *gin.Context) {
	renderTemplate(c, "register.html", gin.H{
		"Title": "Register",
	})
}

//func handleGallery(c *gin.Context) {
//	userID := c.GetUint("user_id")
//	media, err := mediaService.GetUserMedia(userID, media.MediaType(""))
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//		return
//	}
//
//	renderTemplate(c, "gallery.html", gin.H{
//		"Title": "Gallery",
//		"Media": media,
//	})
//}

func handleAlbums(c *gin.Context) {
	userID := c.GetUint("user_id")
	albums, err := albumService.ListUserAlbums(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	renderTemplate(c, "albums.html", gin.H{
		"Title":  "Albums",
		"Albums": albums,
	})
}

func handleLogin(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	token, err := authService.LoginUser(username, password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	c.SetCookie("auth_token", token, 3600*24, "/", "", false, true)

	if c.GetHeader("HX-Request") == "true" {
		c.Header("HX-Redirect", "/")
	} else {
		c.Redirect(http.StatusFound, "/")
	}
}

func handleRegister(c *gin.Context) {
	username := c.PostForm("username")
	email := c.PostForm("email")
	password := c.PostForm("password")

	if err := authService.RegisterUser(username, email, password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if c.GetHeader("HX-Request") == "true" {
		c.Header("HX-Redirect", "/login")
	} else {
		c.Redirect(http.StatusFound, "/login")
	}
}

func handleLogout(c *gin.Context) {
	c.SetCookie("auth_token", "", -1, "/", "", false, true)
	c.Redirect(http.StatusFound, "/login")
}

func handleHome(c *gin.Context) {
	userID := c.GetUint("user_id")
	recentMedia, _ := searchService.GetRecentMedia(userID, 6)

	renderTemplate(c, "home.html", gin.H{
		"Title":       "Home",
		"RecentMedia": recentMedia,
	})
}

func handleUpload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}

	userID := c.GetUint("user_id")
	mediaItem, err := mediaService.ProcessUpload(file, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if c.GetHeader("HX-Request") == "true" {
		renderTemplate(c, "media-item.html", gin.H{"Media": mediaItem})
	} else {
		c.JSON(http.StatusOK, mediaItem)
	}
}

//func handleGetMedia(c *gin.Context) {
//	//id := parseUint(c.Param("id"))
//	userID := c.GetUint("user_id")
//
//	media, err := mediaService.GetUserMedia(userID, "")
//	if err != nil {
//		c.JSON(http.StatusNotFound, gin.H{"error": "Media not found"})
//		return
//	}
//
//	c.JSON(http.StatusOK, media)
//}

//func handleDeleteMedia(c *gin.Context) {
//	id := parseUint(c.Param("id"))
//	userID := c.GetUint("user_id")
//
//	if err := mediaService.DeleteMedia(id, userID); err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//		return
//	}
//
//	c.Status(http.StatusOK)
//}

func handleCreateAlbum(c *gin.Context) {
	name := c.PostForm("name")
	description := c.PostForm("description")
	userID := c.GetUint("user_id")

	album, err := albumService.CreateAlbum(name, description, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if c.GetHeader("HX-Request") == "true" {
		renderTemplate(c, "album-item.html", gin.H{"Album": album})
	} else {
		c.JSON(http.StatusOK, album)
	}
}

func handleGetAlbum(c *gin.Context) {
	id := parseUint(c.Param("id"))
	userID := c.GetUint("user_id")

	album, err := albumService.GetAlbum(id, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Album not found"})
		return
	}

	c.JSON(http.StatusOK, album)
}

func handleUpdateAlbum(c *gin.Context) {
	id := parseUint(c.Param("id"))
	userID := c.GetUint("user_id")
	name := c.PostForm("name")
	description := c.PostForm("description")

	if err := albumService.UpdateAlbum(id, name, description, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func handleDeleteAlbum(c *gin.Context) {
	id := parseUint(c.Param("id"))
	userID := c.GetUint("user_id")

	if err := albumService.DeleteAlbum(id, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func handleAddMediaToAlbum(c *gin.Context) {
	albumID := parseUint(c.Param("id"))
	userID := c.GetUint("user_id")
	var mediaIDs []uint
	if err := c.ShouldBindJSON(&mediaIDs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid media IDs"})
		return
	}

	if err := albumService.AddMediaToAlbum(albumID, mediaIDs, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func handleRemoveMediaFromAlbum(c *gin.Context) {
	albumID := parseUint(c.Param("id"))
	mediaID := parseUint(c.Param("mediaId"))
	userID := c.GetUint("user_id")

	if err := albumService.RemoveMediaFromAlbum(albumID, mediaID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func handleSearch(c *gin.Context) {
	userID := c.GetUint("user_id")
	filter := search.SearchFilter{
		Query:      c.Query("q"),
		MediaType:  c.Query("type"),
		PageSize:   parseIntWithDefault(c.Query("limit"), 20),
		PageNumber: parseIntWithDefault(c.Query("page"), 1),
		SortBy:     c.DefaultQuery("sort", "created_at"),
		SortOrder:  c.DefaultQuery("order", "DESC"),
	}

	results, err := searchService.Search(userID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if c.GetHeader("HX-Request") == "true" {
		renderTemplate(c, "search-results.html", gin.H{"Results": results})
	} else {
		c.JSON(http.StatusOK, results)
	}
}

func handleGetTags(c *gin.Context) {
	userID := c.GetUint("user_id")
	tags, err := searchService.GetPopularTags(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tags": tags})
}

// Helper functions
func renderTemplate(c *gin.Context, name string, data gin.H) {
	if data == nil {
		data = gin.H{}
	}
	data["IsHTMX"] = c.GetHeader("HX-Request") == "true"

	if data["IsHTMX"].(bool) {
		templates.ExecuteTemplate(c.Writer, name, data)
	} else {
		templates.ExecuteTemplate(c.Writer, "layout.html", data)
	}
}

func parseUint(s string) uint {
	n, _ := strconv.ParseUint(s, 10, 32)
	return uint(n)
}

func parseIntWithDefault(s string, def int) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}
