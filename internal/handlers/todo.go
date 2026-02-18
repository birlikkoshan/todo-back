package handlers

import (
	"net/http"
	"strconv"
	"time"

	dom "Worker/internal/domain"
	"Worker/internal/dto"
	"Worker/internal/service"

	"github.com/gin-gonic/gin"
)

type TodoHandler struct {
	svc *service.TodoService
}

func NewTodoHandler(svc *service.TodoService) *TodoHandler {
	return &TodoHandler{svc: svc}
}

// Create godoc
// @Summary      Create a todo
// @Tags         todos
// @Accept       json
// @Produce      json
// @Param        body  body      dto.CreateTodoRequest  true  "Todo body"
// @Success      201   {object}  dto.TodoResponse
// @Failure      400   {object}  map[string]string
// @Router       /todos [post]
func (h *TodoHandler) Create(c *gin.Context) {
	var req dto.CreateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	t, err := h.svc.Create(c.Request.Context(), req.Title, req.Description, req.DueAt.Ptr())
	if err != nil {
		if err == service.ErrInvalidDueDate {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, todoToResponse(t))
}

// List godoc
// @Summary      List all todos
// @Tags         todos
// @Produce      json
// @Security     CookieAuth
// @Success      200  {object}  dto.ListTodosResponse
// @Failure      500  {object}  map[string]string
// @Router       /todos [get]
func (h *TodoHandler) List(c *gin.Context) {
	list, err := h.svc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.ListTodosResponse{Items: todosToResponses(list)})
}

// GetByID godoc
// @Summary      Get a todo by ID
// @Tags         todos
// @Produce      json
// @Security     CookieAuth
// @Param        id   path      int  true  "Todo ID"
// @Success      200  {object}  dto.TodoResponse
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /todos/{id} [get]
func (h *TodoHandler) GetByID(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	t, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == service.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, todoToResponse(t))
}

// Update godoc
// @Summary      Update a todo
// @Tags         todos
// @Accept       json
// @Produce      json
// @Security     CookieAuth
// @Param        id    path      int  true  "Todo ID"
// @Param        body  body      dto.UpdateTodoRequest  true  "Partial update"
// @Success      200   {object}  dto.TodoResponse
// @Failure      400   {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /todos/{id} [patch]
func (h *TodoHandler) Update(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var duePtr *time.Time
	if req.DueAt != nil {
		duePtr = req.DueAt.Ptr()
	}
	t, err := h.svc.Update(c.Request.Context(), id, req.Title, req.Description, duePtr)
	if err != nil {
		if err == service.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		if err == service.ErrInvalidDueDate {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, todoToResponse(t))
}

// Delete godoc
// @Summary      Delete a todo
// @Tags         todos
// @Security     CookieAuth
// @Param        id   path  int  true  "Todo ID"
// @Success      204
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /todos/{id} [delete]
func (h *TodoHandler) Delete(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	err := h.svc.Delete(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// Complete godoc
// @Summary      Mark a todo as done
// @Tags         todos
// @Produce      json
// @Security     CookieAuth
// @Param        id   path      int  true  "Todo ID"
// @Success      200  {object}  dto.TodoResponse
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /todos/{id}/complete [post]
func (h *TodoHandler) Complete(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	t, err := h.svc.Complete(c.Request.Context(), id)
	if err != nil {
		if err == service.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, todoToResponse(t))
}

// Search godoc
// @Summary      Search todos by query
// @Tags         todos
// @Produce      json
// @Security     CookieAuth
// @Param        q    query     string  true  "Search query (title/description)"
// @Success      200  {object}  dto.ListTodosResponse
// @Failure      500  {object}  map[string]string
// @Router       /todos/search [get]
func (h *TodoHandler) Search(c *gin.Context) {
	q := c.Query("q")
	list, err := h.svc.Search(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.ListTodosResponse{Items: todosToResponses(list)})
}

// Overdue godoc
// @Summary      List overdue todos
// @Tags         todos
// @Produce      json
// @Security     CookieAuth
// @Success      200  {object}  dto.ListTodosResponse
// @Failure      500  {object}  map[string]string
// @Router       /todos/overdue [get]
func (h *TodoHandler) Overdue(c *gin.Context) {
	list, err := h.svc.Overdue(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.ListTodosResponse{Items: todosToResponses(list)})
}

func parseID(c *gin.Context, name string) (int64, bool) {
	raw := c.Param(name)
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return 0, false
	}
	return id, true
}

func todoToResponse(t dom.Todo) dto.TodoResponse {
	return dto.TodoResponse{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		IsDone:      t.IsDone,
		DueAt:       t.DueAt,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func todosToResponses(list []dom.Todo) []dto.TodoResponse {
	out := make([]dto.TodoResponse, len(list))
	for i := range list {
		out[i] = todoToResponse(list[i])
	}
	return out
}
