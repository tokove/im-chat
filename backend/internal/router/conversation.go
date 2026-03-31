package router

import (
	"backend/internal/middleware"
	"backend/internal/model"
	"backend/pkg/response"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func handleGetPrivate(c *gin.Context) {
	idStr := c.Param("private_id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		log.Printf("parseInt failed, idStr: %v, err: %v", idStr, err)
		response.JSON(c, http.StatusBadRequest, false, "invalid private_id", nil)
		return
	}

	private, err := model.GetPrivateByID(id)
	if err != nil {
		log.Printf("model.GetPrivateByID failed, err: %v", err)
		response.JSON(c, http.StatusInternalServerError, false, "failed to fetch private by id", nil)
		return
	}

	response.JSON(c, http.StatusOK, true, "success", private)
}

func handleCreatePrivate(c *gin.Context) {
	userIDAny, exists := c.Get(middleware.CtxUserID)
	if !exists {
		log.Printf("handleCreatePrivate: not login")
		response.JSON(c, http.StatusBadRequest, false, "invalid request", nil)
		return
	}
	userID, ok := userIDAny.(int64)
	if !ok {
		log.Printf("handleCreatePrivate: userID type error")
		response.JSON(c, http.StatusBadRequest, false, "invalid request", nil)
		return
	}

	var req struct {
		ReceiverID int64 `json:"receiver_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("handleCreatePrivate: ShouldBindJSON, err: %v", err)
		response.JSON(c, http.StatusBadRequest, false, "invalid request", nil)
		return
	}

	if req.ReceiverID == 0 {
		log.Print("handleCreatePrivate: receiverID is zero")
		response.JSON(c, http.StatusBadRequest, false, "invalid request", nil)
		return
	}

	private, err := model.GetPrivateByUsers(userID, req.ReceiverID)
	if err != nil {
		log.Printf("handleCreatePrivate: GetPrivateByUsers err: %v", err)
		response.JSON(c, http.StatusInternalServerError, false, err.Error(), nil)
		return
	}

	if private == nil {
		private, err = model.CreatePrivate(userID, req.ReceiverID)
		if err != nil {
			log.Printf("handleCreatePrivate: CreatePrivate, err: %v", err)
			response.JSON(c, http.StatusInternalServerError, false, "failed to create private", nil)
			return
		}
		response.JSON(c, http.StatusCreated, true, "success", private)
		return
	}

	response.JSON(c, http.StatusOK, true, "success", private)
}

func handleGetConversations(c *gin.Context) {
	idAny, exist := c.Get("userID")
	if !exist {
		log.Print("handleGetConversations: not login")
		response.JSON(c, http.StatusBadRequest, false, "invalid credentials", nil)
		return
	}

	id, ok := idAny.(int64)
	if !ok {
		log.Print("handleGetConversations: id type error")
		response.JSON(c, http.StatusBadRequest, false, "invalid credentials", nil)
		return
	}

	privates, err := model.GetPrivatesForUser(id)
	if err != nil {
		log.Printf("handleGetConversations: GetPrivatesForUser, err: %v", err)
		response.JSON(c, http.StatusInternalServerError, false, err.Error(), nil)
		return
	}

	response.JSON(c, http.StatusOK, true, "success", gin.H{
		"privates": privates,
	})
}

func handleGetPrivateMessages(c *gin.Context) {
	idStr := c.Param("private_id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		log.Printf("parseInt failed, idStr: %v, err: %v", idStr, err)
		response.JSON(c, http.StatusBadRequest, false, "invalid private_id", nil)
		return
	}

	// http://127.0.0.1/api/conversations/privates/:private_id/messages?page=2&limit=12
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	messages, err := model.GetMessagesByPrivateID(id, page, limit+1)
	if err != nil {
		log.Printf("handleGetPrivateMessages: GetMessagesByPrivateID err: %v", err)
		response.JSON(c, http.StatusInternalServerError, false, "failed to fetch private messages", nil)
		return
	}

	// suit 界面底部显示的是 [1] [2] [3] ... [下一页] 这种按钮
	hasNextPage := false
	if len(messages) > limit {
		hasNextPage = true
		messages = messages[:limit]
	}

	response.JSON(c, http.StatusOK, true, "success", gin.H{
		"messages":      messages,
		"page":          page,
		"limit":         limit,
		"has_next_page": hasNextPage, 
	})
}
