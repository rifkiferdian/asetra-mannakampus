package controllers

import (
	"fmt"
	"gobase-app/config"
	"gobase-app/models"
	"gobase-app/repositories"
	"gobase-app/services"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func PurchaseRequestIndex(c *gin.Context) {
	service := buildPurchaseRequestService()
	items, err := service.GetPurchaseRequests()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	storeRepo := &repositories.StoreRepository{DB: config.DB}
	stores, err := storeRepo.GetAll()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	divRepo := &repositories.DivisionRepository{DB: config.DB}
	divisions, err := divRepo.GetAll()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	Render(c, "purchase_request.html", gin.H{
		"Title":            "Purchase Requests",
		"Page":             "purchase_request",
		"PurchaseRequests": items,
		"Stores":           stores,
		"Divisions":        divisions,
	})
}

func PurchaseRequestFormIndex(c *gin.Context) {
	renderPurchaseRequestForm(c, "", nil)
}

func PurchaseRequestDetailIndex(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.String(http.StatusBadRequest, "purchase request tidak valid")
		return
	}

	session := sessions.Default(c)
	userID := sessionUserID(session)

	service := buildPurchaseRequestService()
	detail, err := service.GetPurchaseRequestDetail(id, userID)
	if err != nil {
		c.String(http.StatusNotFound, err.Error())
		return
	}

	Render(c, "purchase_request_detail.html", gin.H{
		"Title": "PR Detail",
		"Page":  "purchase_request",
		"PR":    detail,
	})
}

func PurchaseRequestStore(c *gin.Context) {
	input, cleanupPaths, errMessage := bindPurchaseRequestInput(c)
	if errMessage != "" {
		renderPurchaseRequestForm(c, errMessage, nil)
		return
	}

	service := buildPurchaseRequestService()
	if err := service.CreatePurchaseRequest(input); err != nil {
		cleanupUploadedFiles(cleanupPaths)
		renderPurchaseRequestForm(c, err.Error(), nil)
		return
	}

	c.Redirect(http.StatusSeeOther, "/purchase-requests")
}

func renderPurchaseRequestForm(c *gin.Context, message string, persisted gin.H) {
	storeRepo := &repositories.StoreRepository{DB: config.DB}
	allStores, err := storeRepo.GetAll()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	session := sessions.Default(c)
	allowedStoreIDs := parseSessionStoreIDs(session)
	stores := allStores
	if len(allowedStoreIDs) > 0 {
		filtered, err := storeRepo.GetByIDs(allowedStoreIDs)
		if err == nil && len(filtered) > 0 {
			stores = filtered
		}
	}

	divRepo := &repositories.DivisionRepository{DB: config.DB}
	divisions, err := divRepo.GetAll()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	glRepo := &repositories.GLAccountRepository{DB: config.DB}
	glAccounts, err := glRepo.GetAll()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	data := gin.H{
		"Title":      "Form Purchase Request",
		"Page":       "purchase_request_form",
		"Stores":     stores,
		"Divisions":  divisions,
		"GLAccounts": glAccounts,
		"Error":      message,
	}
	if persisted != nil {
		for key, value := range persisted {
			data[key] = value
		}
	}

	Render(c, "purchase_request_form.html", data)
}

func buildPurchaseRequestService() *services.PurchaseRequestService {
	return &services.PurchaseRequestService{
		Repo:         &repositories.PurchaseRequestRepository{DB: config.DB},
		DivisionRepo: &repositories.DivisionRepository{DB: config.DB},
		StoreRepo:    &repositories.StoreRepository{DB: config.DB},
		GlRepo:       &repositories.GLAccountRepository{DB: config.DB},
	}
}

func bindPurchaseRequestInput(c *gin.Context) (models.PurchaseRequestCreateInput, []string, string) {
	session := sessions.Default(c)
	userID := sessionUserID(session)
	if userID <= 0 {
		return models.PurchaseRequestCreateInput{}, nil, "user login tidak valid"
	}

	storeID, err := strconv.Atoi(strings.TrimSpace(c.PostForm("store_id")))
	if err != nil {
		return models.PurchaseRequestCreateInput{}, nil, "store wajib dipilih"
	}

	divisionID := 0
	if val := strings.TrimSpace(c.PostForm("division_id")); val != "" {
		divisionID, err = strconv.Atoi(val)
		if err != nil {
			return models.PurchaseRequestCreateInput{}, nil, "division tidak valid"
		}
	}

	glAccountID, err := strconv.Atoi(strings.TrimSpace(c.PostForm("gl_account_id")))
	if err != nil {
		return models.PurchaseRequestCreateInput{}, nil, "GL account wajib dipilih"
	}

	itemNames := c.PostFormArray("item_name[]")
	qtyVals := c.PostFormArray("qty[]")
	uoms := c.PostFormArray("uom[]")
	priceVals := c.PostFormArray("est_unit_price[]")
	notesVals := c.PostFormArray("notes[]")

	if len(itemNames) == 0 || len(itemNames) != len(qtyVals) || len(itemNames) != len(uoms) || len(itemNames) != len(priceVals) || len(itemNames) != len(notesVals) {
		return models.PurchaseRequestCreateInput{}, nil, "data item PR tidak lengkap"
	}

	items := make([]models.PurchaseRequestItemInput, 0, len(itemNames))
	for i := range itemNames {
		qty, err := strconv.ParseFloat(strings.TrimSpace(qtyVals[i]), 64)
		if err != nil {
			return models.PurchaseRequestCreateInput{}, nil, fmt.Sprintf("qty item baris %d tidak valid", i+1)
		}
		price, err := strconv.ParseFloat(strings.TrimSpace(priceVals[i]), 64)
		if err != nil {
			return models.PurchaseRequestCreateInput{}, nil, fmt.Sprintf("estimasi harga item baris %d tidak valid", i+1)
		}
		item := models.PurchaseRequestItemInput{
			ItemName:     strings.TrimSpace(itemNames[i]),
			Qty:          qty,
			UOM:          strings.TrimSpace(uoms[i]),
			EstUnitPrice: price,
			Notes:        strings.TrimSpace(notesVals[i]),
		}
		items = append(items, item)
	}

	attachments, cleanupPaths, err := savePRAttachments(c)
	if err != nil {
		return models.PurchaseRequestCreateInput{}, cleanupPaths, err.Error()
	}

	input := models.PurchaseRequestCreateInput{
		RequesterUserID: userID,
		StoreID:         storeID,
		DivisionID:      divisionID,
		GLAccountID:     glAccountID,
		SpendType:       strings.ToUpper(strings.TrimSpace(c.PostForm("spend_type"))),
		UrgentLevel:     strings.ToUpper(strings.TrimSpace(c.PostForm("urgent_level"))),
		NeededDate:      strings.TrimSpace(c.PostForm("needed_date")),
		Justification:   strings.TrimSpace(c.PostForm("justification")),
		Action:          strings.ToLower(strings.TrimSpace(c.PostForm("action"))),
		Items:           items,
		Attachments:     attachments,
		AuditContext: models.AuditContext{
			ActorUserID: userID,
			IPAddress:   c.ClientIP(),
			UserAgent:   c.Request.UserAgent(),
		},
	}

	return input, cleanupPaths, ""
}

func savePRAttachments(c *gin.Context) ([]models.AttachmentFileInput, []string, error) {
	form, err := c.MultipartForm()
	if err != nil {
		if err == http.ErrNotMultipart {
			return nil, nil, nil
		}
		// Gin may return "request Content-Type isn't multipart/form-data"
		if strings.Contains(strings.ToLower(err.Error()), "multipart") {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	if form == nil || len(form.File["attachments"]) == 0 {
		return nil, nil, nil
	}

	now := time.Now()
	targetDir := filepath.Join("storage", "attachments", "pr", now.Format("2006"), now.Format("01"))
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return nil, nil, err
	}

	attachments := make([]models.AttachmentFileInput, 0, len(form.File["attachments"]))
	cleanupPaths := make([]string, 0, len(form.File["attachments"]))

	for _, file := range form.File["attachments"] {
		if file == nil || strings.TrimSpace(file.Filename) == "" {
			continue
		}
		safeName := sanitizeFileName(file.Filename)
		fileName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), safeName)
		fullPath := filepath.Join(targetDir, fileName)
		if err := c.SaveUploadedFile(file, fullPath); err != nil {
			cleanupUploadedFiles(cleanupPaths)
			return nil, cleanupPaths, err
		}
		cleanupPaths = append(cleanupPaths, fullPath)

		mimeType := detectMimeType(file)
		attachments = append(attachments, models.AttachmentFileInput{
			FileName: safeName,
			FilePath: filepath.ToSlash(fullPath),
			MimeType: mimeType,
			FileSize: file.Size,
		})
	}

	return attachments, cleanupPaths, nil
}

func detectMimeType(file *multipart.FileHeader) string {
	if file == nil {
		return ""
	}
	if file.Header != nil {
		return file.Header.Get("Content-Type")
	}
	return ""
}

func sanitizeFileName(name string) string {
	name = filepath.Base(strings.TrimSpace(name))
	name = strings.ReplaceAll(name, " ", "_")
	return name
}

func cleanupUploadedFiles(paths []string) {
	for _, path := range paths {
		if strings.TrimSpace(path) == "" {
			continue
		}
		_ = os.Remove(path)
	}
}

func parseSessionStoreIDs(session sessions.Session) []int {
	raw := session.Get("user")
	if raw == nil {
		return nil
	}
	switch val := raw.(type) {
	case models.SessionUser:
		return splitCSVToInts(val.StoreID)
	case map[string]interface{}:
		if storeID, ok := val["store_id"].(string); ok {
			return splitCSVToInts(storeID)
		}
		if storeID, ok := val["StoreID"].(string); ok {
			return splitCSVToInts(storeID)
		}
	case gin.H:
		if storeID, ok := val["store_id"].(string); ok {
			return splitCSVToInts(storeID)
		}
		if storeID, ok := val["StoreID"].(string); ok {
			return splitCSVToInts(storeID)
		}
	}
	return nil
}

func splitCSVToInts(value string) []int {
	parts := strings.Split(strings.TrimSpace(value), ",")
	var ids []int
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		id, err := strconv.Atoi(part)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}
	return ids
}

func sessionUserID(session sessions.Session) int {
	if value := session.Get("user_id"); value != nil {
		switch id := value.(type) {
		case int:
			return id
		case int64:
			return int(id)
		case float64:
			return int(id)
		}
	}

	if raw := session.Get("user"); raw != nil {
		switch val := raw.(type) {
		case models.SessionUser:
			return val.UserID
		case map[string]interface{}:
			if value, ok := val["user_id"]; ok {
				return normalizeSessionUserID(value)
			}
			if value, ok := val["UserID"]; ok {
				return normalizeSessionUserID(value)
			}
		case gin.H:
			if value, ok := val["user_id"]; ok {
				return normalizeSessionUserID(value)
			}
			if value, ok := val["UserID"]; ok {
				return normalizeSessionUserID(value)
			}
		}
	}

	return 0
}

func normalizeSessionUserID(value interface{}) int {
	switch id := value.(type) {
	case int:
		return id
	case int64:
		return int(id)
	case float64:
		return int(id)
	default:
		return 0
	}
}
