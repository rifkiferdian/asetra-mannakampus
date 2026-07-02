package controllers

import (
	"fmt"
	"gobase-app/config"
	"gobase-app/models"
	"gobase-app/repositories"
	"gobase-app/services"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func PurchaseOrderIndex(c *gin.Context) {
	service := buildPurchaseOrderService()

	orders, err := service.GetPurchaseOrders()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	readyPRs, err := service.GetApprovedPRReadyForPO()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	stats := buildPOStats(orders, readyPRs)

	Render(c, "purchase_order.html", gin.H{
		"Title":          "Purchase Orders",
		"Page":           "purchase_order",
		"PurchaseOrders": orders,
		"ReadyPRs":       readyPRs,
		"Stats":          stats,
		"Success":        strings.TrimSpace(c.Query("success")),
		"Error":          strings.TrimSpace(c.Query("error")),
	})
}

func PurchaseOrderCreateFromPR(c *gin.Context) {
	prID, err := strconv.ParseInt(c.Param("pr_id"), 10, 64)
	if err != nil || prID <= 0 {
		c.Redirect(http.StatusSeeOther, "/purchase-orders?error="+url.QueryEscape("purchase request tidak valid"))
		return
	}

	session := sessions.Default(c)
	userID := sessionUserID(session)

	service := buildPurchaseOrderService()
	form, err := service.GetCreateForm(prID, userID)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/purchase-orders?error="+url.QueryEscape(err.Error()))
		return
	}

	Render(c, "purchase_order_form.html", gin.H{
		"Title": "Create PO",
		"Page":  "purchase_order",
		"Form":  form,
		"Error": strings.TrimSpace(c.Query("error")),
	})
}

func PurchaseOrderStore(c *gin.Context) {
	input, errMessage := bindPurchaseOrderCreateInput(c)
	if errMessage != "" {
		c.Redirect(http.StatusSeeOther, "/purchase-orders/create-from-pr/"+strconv.FormatInt(input.PRID, 10)+"?error="+url.QueryEscape(errMessage))
		return
	}

	service := buildPurchaseOrderService()
	poID, err := service.CreateFromPR(input)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/purchase-orders/create-from-pr/"+strconv.FormatInt(input.PRID, 10)+"?error="+url.QueryEscape(err.Error()))
		return
	}

	c.Redirect(http.StatusSeeOther, "/purchase-orders/"+strconv.FormatInt(poID, 10)+"?success="+url.QueryEscape("Purchase order berhasil dibuat"))
}

func PurchaseOrderDetailIndex(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.Redirect(http.StatusSeeOther, "/purchase-orders?error="+url.QueryEscape("purchase order tidak valid"))
		return
	}

	service := buildPurchaseOrderService()
	detail, err := service.GetPurchaseOrderDetail(id)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/purchase-orders?error="+url.QueryEscape(err.Error()))
		return
	}

	Render(c, "purchase_order_detail.html", gin.H{
		"Title":   "PO Detail",
		"Page":    "purchase_order",
		"PO":      detail,
		"Success": strings.TrimSpace(c.Query("success")),
		"Error":   strings.TrimSpace(c.Query("error")),
	})
}

func bindPurchaseOrderCreateInput(c *gin.Context) (models.PurchaseOrderCreateInput, string) {
	session := sessions.Default(c)
	userID := sessionUserID(session)

	prID, err := strconv.ParseInt(strings.TrimSpace(c.PostForm("pr_id")), 10, 64)
	if err != nil || prID <= 0 {
		return models.PurchaseOrderCreateInput{}, "purchase request tidak valid"
	}

	vendorID, err := strconv.ParseInt(strings.TrimSpace(c.PostForm("vendor_id")), 10, 64)
	if err != nil || vendorID <= 0 {
		return models.PurchaseOrderCreateInput{PRID: prID}, "vendor wajib dipilih"
	}

	prItemIDs := c.PostFormArray("pr_item_id[]")
	unitPrices := c.PostFormArray("unit_price[]")
	if len(prItemIDs) == 0 || len(prItemIDs) != len(unitPrices) {
		return models.PurchaseOrderCreateInput{PRID: prID}, "data item PO tidak lengkap"
	}

	items := make([]models.PurchaseOrderItemInput, 0, len(prItemIDs))
	for i := range prItemIDs {
		prItemID, err := strconv.ParseInt(strings.TrimSpace(prItemIDs[i]), 10, 64)
		if err != nil || prItemID <= 0 {
			return models.PurchaseOrderCreateInput{PRID: prID}, fmt.Sprintf("item PR baris %d tidak valid", i+1)
		}

		unitPrice, err := strconv.ParseFloat(strings.TrimSpace(unitPrices[i]), 64)
		if err != nil {
			return models.PurchaseOrderCreateInput{PRID: prID}, fmt.Sprintf("harga final baris %d tidak valid", i+1)
		}

		items = append(items, models.PurchaseOrderItemInput{
			PRItemID:  prItemID,
			UnitPrice: unitPrice,
		})
	}

	return models.PurchaseOrderCreateInput{
		PRID:     prID,
		VendorID: vendorID,
		Items:    items,
		AuditContext: models.AuditContext{
			ActorUserID: userID,
			IPAddress:   c.ClientIP(),
			UserAgent:   c.Request.UserAgent(),
		},
	}, ""
}

func buildPurchaseOrderService() *services.PurchaseOrderService {
	return &services.PurchaseOrderService{
		Repo:       &repositories.PurchaseOrderRepository{DB: config.DB},
		VendorRepo: &repositories.VendorRepository{DB: config.DB},
	}
}

func buildPOStats(orders []models.PurchaseOrder, readyPRs []models.ApprovedPRForPO) gin.H {
	stats := gin.H{
		"total":     len(orders),
		"ready":     len(readyPRs),
		"approved":  0,
		"receiving": 0,
		"closed":    0,
	}
	for _, order := range orders {
		switch order.Status {
		case "APPROVED":
			stats["approved"] = stats["approved"].(int) + 1
		case "RECEIVING":
			stats["receiving"] = stats["receiving"].(int) + 1
		case "CLOSED":
			stats["closed"] = stats["closed"].(int) + 1
		}
	}
	return stats
}
