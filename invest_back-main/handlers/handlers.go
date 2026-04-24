package handlers

import (
	"bytes"
	"fmt"
	"io"
	"investtrack-backend/database"
	"investtrack-backend/models"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ───────────────── FILE UPLOAD ─────────────────

func UploadImage(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowed := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true,
		".mp4": true, ".webm": true, ".mov": true, ".avi": true, ".mkv": true,
	}
	if !allowed[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Faqat rasm (jpg, png, webp, gif) yoki video (mp4, webm, mov, avi, mkv) fayllar qabul qilinadi"})
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_SERVICE_KEY")

	if supabaseURL == "" || supabaseKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "SUPABASE_URL va SUPABASE_SERVICE_KEY environment variable sozlanmagan"})
		return
	}

	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	uploadURL := fmt.Sprintf("%s/storage/v1/object/investtrack/%s", supabaseURL, filename)

	req, err := http.NewRequest("POST", uploadURL, bytes.NewReader(data))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Request creation failed"})
		return
	}
	req.Header.Set("Authorization", "Bearer "+supabaseKey)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-upsert", "true")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload request failed: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		errMsg := fmt.Sprintf("Supabase upload failed (status=%d): %s", resp.StatusCode, string(body))
		fmt.Println("UPLOAD ERROR:", errMsg)
		c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		return
	}

	publicURL := fmt.Sprintf("%s/storage/v1/object/public/investtrack/%s", supabaseURL, filename)

	c.JSON(http.StatusOK, gin.H{
		"url":      publicURL,
		"filename": filename,
	})
}

// ───────────────── OBJECTS ─────────────────

func GetTopObjects(c *gin.Context) {
	var objects []models.InvestmentObject
	database.DB.Preload("Owner").
		Where("top_rating >= ?", 4.0).
		Order("top_rating DESC").
		Limit(10).
		Find(&objects)
	enrichObjects(objects)
	c.JSON(http.StatusOK, objects)
}

func GetObjects(c *gin.Context) {
	var objects []models.InvestmentObject
	query := database.DB.Preload("Owner")

	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if objType := c.Query("type"); objType != "" {
		query = query.Where("type = ?", objType)
	}
	if featured := c.Query("featured"); featured == "true" {
		query = query.Where("featured = ?", true)
	}
	if topRated := c.Query("topRated"); topRated == "true" {
		query = query.Where("top_rated = ?", true)
	}

	query.Find(&objects)
	enrichObjects(objects)
	c.JSON(http.StatusOK, objects)
}

func GetObject(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var object models.InvestmentObject
	result := database.DB.Preload("Owner").Preload("Investor").Preload("Comments").Preload("Images", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC")
	}).First(&object, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Object not found"})
		return
	}
	enrichObject(&object)
	c.JSON(http.StatusOK, object)
}

func CreateObject(c *gin.Context) {
	var input models.InvestmentObject
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	database.DB.Create(&input)
	c.JSON(http.StatusCreated, input)
}

func UpdateObject(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var object models.InvestmentObject
	if database.DB.First(&object, id).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Object not found"})
		return
	}
	var input models.InvestmentObject
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	database.DB.Model(&object).
		Select("*").
		Omit("id", "created_at", "comments", "images", "average_rating", "comment_count").
		Updates(input)
	database.DB.Preload("Owner").Preload("Investor").First(&object, id)
	c.JSON(http.StatusOK, object)
}

func DeleteObject(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	// Delete dependent records first to avoid FK constraint violations in PostgreSQL
	database.DB.Where("object_id = ?", id).Delete(&models.Comment{})
	database.DB.Where("object_id = ?", id).Delete(&models.ObjectImage{})
	if result := database.DB.Delete(&models.InvestmentObject{}, id); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Deleted successfully"})
}

// enrichObject — single object (used in GetObject)
func enrichObject(o *models.InvestmentObject) {
	type stat struct {
		Count     int64
		AvgRating float64
	}
	var s stat
	database.DB.Model(&models.Comment{}).
		Select("COUNT(*) as count, COALESCE(AVG(CAST(rating AS REAL)), 0) as avg_rating").
		Where("object_id = ?", o.ID).
		Scan(&s)
	o.CommentCount = int(s.Count)
	if s.Count > 0 {
		avg := s.AvgRating
		o.AverageRating = &avg
	}
}

// enrichObjects — batch version, 1 query instead of N*2 queries
func enrichObjects(objects []models.InvestmentObject) {
	if len(objects) == 0 {
		return
	}
	ids := make([]uint, len(objects))
	for i, o := range objects {
		ids[i] = o.ID
	}

	type stat struct {
		ObjectID  uint
		Count     int64
		AvgRating float64
	}
	var stats []stat
	database.DB.Model(&models.Comment{}).
		Select("object_id, COUNT(*) as count, COALESCE(AVG(CAST(rating AS REAL)), 0) as avg_rating").
		Where("object_id IN ?", ids).
		Group("object_id").
		Scan(&stats)

	statMap := make(map[uint]stat, len(stats))
	for _, s := range stats {
		statMap[s.ObjectID] = s
	}

	for i := range objects {
		if s, ok := statMap[objects[i].ID]; ok {
			objects[i].CommentCount = int(s.Count)
			avg := s.AvgRating
			objects[i].AverageRating = &avg
		}
	}
}

// ───────────────── OWNERS ─────────────────

func GetOwners(c *gin.Context) {
	var owners []models.Owner
	database.DB.Find(&owners)
	c.JSON(http.StatusOK, owners)
}

func GetOwner(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var owner models.Owner
	if database.DB.Preload("Objects").First(&owner, id).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Owner not found"})
		return
	}
	c.JSON(http.StatusOK, owner)
}

func CreateOwner(c *gin.Context) {
	var input models.Owner
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	database.DB.Create(&input)
	c.JSON(http.StatusCreated, input)
}

func UpdateOwner(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var owner models.Owner
	if database.DB.First(&owner, id).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Owner not found"})
		return
	}
	var input models.Owner
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	database.DB.Model(&owner).Updates(input)
	c.JSON(http.StatusOK, owner)
}

func DeleteOwner(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	// Nullify owner_id on related objects before deleting
	database.DB.Model(&models.InvestmentObject{}).Where("owner_id = ?", id).Update("owner_id", nil)
	if result := database.DB.Delete(&models.Owner{}, id); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Deleted successfully"})
}

// ───────────────── INVESTORS ─────────────────

func GetInvestors(c *gin.Context) {
	var investors []models.Investor
	database.DB.Find(&investors)
	c.JSON(http.StatusOK, investors)
}

func GetInvestor(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var investor models.Investor
	if database.DB.Preload("Objects").First(&investor, id).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Investor not found"})
		return
	}
	c.JSON(http.StatusOK, investor)
}

func CreateInvestor(c *gin.Context) {
	var input models.Investor
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	database.DB.Create(&input)
	c.JSON(http.StatusCreated, input)
}

func UpdateInvestor(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var investor models.Investor
	if database.DB.First(&investor, id).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Investor not found"})
		return
	}
	var input models.Investor
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	database.DB.Model(&investor).Updates(input)
	c.JSON(http.StatusOK, investor)
}

func DeleteInvestor(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	// Nullify investor_id on related objects before deleting
	database.DB.Model(&models.InvestmentObject{}).Where("investor_id = ?", id).Update("investor_id", nil)
	if result := database.DB.Delete(&models.Investor{}, id); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Deleted successfully"})
}

// ───────────────── OBJECT IMAGES ─────────────────

func GetObjectImages(c *gin.Context) {
	var images []models.ObjectImage
	query := database.DB.Model(&models.ObjectImage{})
	if objectID := c.Query("objectId"); objectID != "" {
		query = query.Where("object_id = ?", objectID)
	}
	query.Order("sort_order ASC").Find(&images)
	c.JSON(http.StatusOK, images)
}

func CreateObjectImage(c *gin.Context) {
	var input models.ObjectImage
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Auto-detect media type from file extension when not provided
	if input.MediaType == "" {
		lower := strings.ToLower(input.ImageURL)
		if strings.HasSuffix(lower, ".mp4") || strings.HasSuffix(lower, ".webm") ||
			strings.HasSuffix(lower, ".mov") || strings.HasSuffix(lower, ".avi") ||
			strings.HasSuffix(lower, ".mkv") {
			input.MediaType = "video"
		} else {
			input.MediaType = "image"
		}
	}

	savedMediaType := input.MediaType

	result := database.DB.Create(&input)
	if result.Error != nil {
		// If the media_type column doesn't exist yet, retry without it
		if strings.Contains(result.Error.Error(), "media_type") {
			input.MediaType = ""
			if retry := database.DB.Omit("MediaType").Create(&input); retry.Error != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": retry.Error.Error()})
				return
			}
			// Try to add the column now so future inserts work
			database.DB.Exec(`ALTER TABLE object_images ADD COLUMN IF NOT EXISTS media_type TEXT DEFAULT 'image'`)
			database.DB.Exec("UPDATE object_images SET media_type = ? WHERE id = ?", savedMediaType, input.ID)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		}
	}

	// Re-fetch to get all DB-generated fields (id, timestamps)
	database.DB.First(&input, input.ID)
	// Restore detected type in response (column may be empty if migration just ran)
	if input.MediaType == "" {
		input.MediaType = savedMediaType
	}
	c.JSON(http.StatusCreated, input)
}

func DeleteObjectImage(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	database.DB.Delete(&models.ObjectImage{}, id)
	c.JSON(http.StatusOK, gin.H{"message": "Deleted successfully"})
}

// ───────────────── COMMENTS ─────────────────

func GetComments(c *gin.Context) {
	var comments []models.Comment
	query := database.DB.Model(&models.Comment{})
	if objectID := c.Query("objectId"); objectID != "" {
		query = query.Where("object_id = ?", objectID)
	}
	query.Find(&comments)
	c.JSON(http.StatusOK, comments)
}

func CreateComment(c *gin.Context) {
	var input models.Comment
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	database.DB.Create(&input)
	c.JSON(http.StatusCreated, input)
}

func DeleteComment(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	database.DB.Delete(&models.Comment{}, id)
	c.JSON(http.StatusOK, gin.H{"message": "Deleted successfully"})
}

// ───────────────── STATISTICS ─────────────────

func GetStatistics(c *gin.Context) {
	var stats []models.Statistic
	database.DB.Order("sort_order ASC").Find(&stats)
	c.JSON(http.StatusOK, stats)
}

func CreateStatistic(c *gin.Context) {
	var input models.Statistic
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	database.DB.Create(&input)
	c.JSON(http.StatusCreated, input)
}

func UpdateStatistic(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var stat models.Statistic
	if database.DB.First(&stat, id).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Statistic not found"})
		return
	}
	var input models.Statistic
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	database.DB.Model(&stat).Updates(input)
	c.JSON(http.StatusOK, stat)
}

func DeleteStatistic(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	database.DB.Delete(&models.Statistic{}, id)
	c.JSON(http.StatusOK, gin.H{"message": "Deleted successfully"})
}

// ───────────────── FAQS ─────────────────

func GetFAQs(c *gin.Context) {
	var faqs []models.FAQ
	database.DB.Order("sort_order ASC").Find(&faqs)
	c.JSON(http.StatusOK, faqs)
}

func CreateFAQ(c *gin.Context) {
	var input models.FAQ
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	database.DB.Create(&input)
	c.JSON(http.StatusCreated, input)
}

func UpdateFAQ(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var faq models.FAQ
	if database.DB.First(&faq, id).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "FAQ not found"})
		return
	}
	var input models.FAQ
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	database.DB.Model(&faq).Updates(input)
	c.JSON(http.StatusOK, faq)
}

func DeleteFAQ(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	database.DB.Delete(&models.FAQ{}, id)
	c.JSON(http.StatusOK, gin.H{"message": "Deleted successfully"})
}

// ───────────────── DASHBOARD ─────────────────

func GetDashboardSummary(c *gin.Context) {
	var summary models.DashboardSummary
	database.DB.Model(&models.InvestmentObject{}).Count(&summary.TotalObjects)
	database.DB.Model(&models.InvestmentObject{}).Where("status = ?", "active").Count(&summary.ActiveObjects)
	database.DB.Model(&models.Owner{}).Count(&summary.TotalOwners)
	database.DB.Model(&models.Comment{}).Count(&summary.TotalComments)
	database.DB.Model(&models.InvestmentObject{}).Where("featured = ?", true).Count(&summary.FeaturedCount)
	database.DB.Model(&models.InvestmentObject{}).Select("COALESCE(SUM(investment_value), 0)").Scan(&summary.TotalValue)
	database.DB.Model(&models.InvestmentObject{}).Select("COALESCE(AVG(return_rate), 0)").Scan(&summary.AvgReturnRate)
	c.JSON(http.StatusOK, summary)
}
