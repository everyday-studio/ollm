package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"

	"strconv"

	"github.com/everyday-studio/ollm/internal/config"
	"github.com/everyday-studio/ollm/internal/domain"
	"github.com/everyday-studio/ollm/internal/middleware"
	"github.com/everyday-studio/ollm/view/admin"
)

// Render is a helper function to render Templ components using Echo.
func Render(c echo.Context, statusCode int, t templ.Component) error {
	c.Response().Writer.WriteHeader(statusCode)
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	return t.Render(c.Request().Context(), c.Response().Writer)
}

type AdminHandler struct {
	userUseCase  domain.UserUseCase
	gameUseCase  domain.GameUseCase
	matchUseCase domain.MatchUseCase
	authUseCase  domain.AuthUsecase
	config       *config.Config
}

func NewAdminHandler(e *echo.Echo, userUseCase domain.UserUseCase, gameUseCase domain.GameUseCase, matchUseCase domain.MatchUseCase, authUseCase domain.AuthUsecase, cfg *config.Config) *AdminHandler {
	handler := &AdminHandler{
		userUseCase:  userUseCase,
		gameUseCase:  gameUseCase,
		matchUseCase: matchUseCase,
		authUseCase:  authUseCase,
		config:       cfg,
	}

	adminPath := cfg.App.AdminPath
	if adminPath == "" {
		adminPath = "/admin"
	}

	publicGroup := e.Group(adminPath)
	publicGroup.GET("/login", handler.LoginForm)
	publicGroup.POST("/login", handler.Login)

	adminGroup := e.Group(adminPath, middleware.AllowRoles(domain.RoleAdmin))
	adminGroup.GET("/dashboard", handler.Dashboard)

	adminGroup.GET("/users", handler.Users)

	adminGroup.GET("/games", handler.Games)
	adminGroup.GET("/games/create", handler.GameCreateForm)
	adminGroup.POST("/games", handler.CreateGame)
	adminGroup.GET("/games/:id/edit", handler.GameEditForm)
	adminGroup.PUT("/games/:id", handler.UpdateGame)
	adminGroup.PATCH("/games/:id/visibility", handler.ToggleGameVisibility)

	return handler
}

func (h *AdminHandler) GameCreateForm(c echo.Context) error {
	adminPath := h.config.App.AdminPath
	if adminPath == "" {
		adminPath = "/admin"
	}

	// Get current admin ID to set as default author
	authorID, _ := c.Get("user_id").(string)

	return Render(c, http.StatusOK, admin.GameCreatePage(adminPath, authorID))
}

func (h *AdminHandler) CreateGame(c echo.Context) error {
	type createGameRequest struct {
		Title          string `json:"title"`
		Description    string `json:"description"`
		AuthorID       string `json:"author_id"`
		SystemPrompt   string `json:"system_prompt"`
		FirstMessage   string `json:"first_message"`
		JudgeType      string `json:"judge_type"`
		JudgeCondition string `json:"judge_condition"`
		MaxTurns       string `json:"max_turns"`
	}

	req := new(createGameRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse(domain.ErrInvalidInput))
	}

	maxTurns, err := strconv.Atoi(req.MaxTurns)
	if err != nil {
		maxTurns = 10
	}

	domainReq := &domain.CreateGameRequest{
		Title:          req.Title,
		Description:    req.Description,
		AuthorID:       req.AuthorID,
		SystemPrompt:   req.SystemPrompt,
		FirstMessage:   req.FirstMessage,
		JudgeType:      domain.JudgeType(req.JudgeType),
		JudgeCondition: req.JudgeCondition,
		MaxTurns:       maxTurns,
	}

	ctx := c.Request().Context()
	_, err = h.gameUseCase.Create(ctx, domainReq)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse(domain.ErrInternal))
	}

	adminPath := h.config.App.AdminPath
	if adminPath == "" {
		adminPath = "/admin"
	}

	// HX-Redirect to games list on success
	c.Response().Header().Set("HX-Redirect", adminPath+"/games")
	return c.NoContent(http.StatusCreated)
}

func (h *AdminHandler) GameEditForm(c echo.Context) error {
	adminPath := h.config.App.AdminPath
	if adminPath == "" {
		adminPath = "/admin"
	}

	id := c.Param("id")
	ctx := c.Request().Context()
	game, err := h.gameUseCase.GetByID(ctx, id)
	if err != nil {
		return c.Redirect(http.StatusFound, adminPath+"/games")
	}

	return Render(c, http.StatusOK, admin.GameEditPage(adminPath, *game))
}

func (h *AdminHandler) UpdateGame(c echo.Context) error {
	id := c.Param("id")

	type updateGameRequest struct {
		Title          string `json:"title"`
		Description    string `json:"description"`
		SystemPrompt   string `json:"system_prompt"`
		FirstMessage   string `json:"first_message"`
		JudgeType      string `json:"judge_type"`
		JudgeCondition string `json:"judge_condition"`
		MaxTurns       string `json:"max_turns"`
	}

	req := new(updateGameRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse(domain.ErrInvalidInput))
	}

	// Make sure fields are actually provided
	if req.Title == "" || req.SystemPrompt == "" {
		return c.JSON(http.StatusBadRequest, ErrResponse(domain.ErrInvalidInput))
	}

	maxTurns, err := strconv.Atoi(req.MaxTurns)
	if err != nil {
		maxTurns = 10
	}

	judgeType := domain.JudgeType(req.JudgeType)

	domainReq := &domain.UpdateGameRequest{
		Title:          &req.Title,
		Description:    &req.Description,
		SystemPrompt:   &req.SystemPrompt,
		FirstMessage:   &req.FirstMessage,
		JudgeType:      &judgeType,
		JudgeCondition: &req.JudgeCondition,
		MaxTurns:       &maxTurns,
	}

	ctx := c.Request().Context()
	_, err = h.gameUseCase.Update(ctx, id, domainReq)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse(domain.ErrInternal))
	}

	adminPath := h.config.App.AdminPath
	if adminPath == "" {
		adminPath = "/admin"
	}

	// HX-Redirect to games list on success
	c.Response().Header().Set("HX-Redirect", adminPath+"/games")
	return c.NoContent(http.StatusOK)
}

func (h *AdminHandler) LoginForm(c echo.Context) error {
	adminPath := h.config.App.AdminPath
	if adminPath == "" {
		adminPath = "/admin"
	}
	return Render(c, http.StatusOK, admin.Login(adminPath))
}

func (h *AdminHandler) Login(c echo.Context) error {
	req := new(domain.LoginRequest)
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrResponse(domain.ErrInvalidInput))
	}

	if req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, ErrResponse(domain.ErrInvalidInput))
	}

	ctx := c.Request().Context()
	loginResponse, err := h.authUseCase.Login(ctx, req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUnauthorized):
			return c.JSON(http.StatusUnauthorized, ErrResponse(domain.ErrUnauthorized))
		default:
			return c.JSON(http.StatusInternalServerError, ErrResponse(domain.ErrInternal))
		}
	}

	// For extra security, verify if user is really admin
	user, getErr := h.userUseCase.GetByID(ctx, loginResponse.ID)
	if getErr != nil || user.Role != domain.RoleAdmin {
		return c.JSON(http.StatusForbidden, ErrResponse(domain.ErrForbidden))
	}

	// Create Access Token Cookie
	accessCookie := &http.Cookie{
		Name:     "access_token",
		Value:    loginResponse.AccessToken,
		Path:     "/",
		Domain:   h.config.Secure.JWT.Cookie.Domain,
		MaxAge:   3600, // 1 hour for now, standard access token lifespan
		Secure:   h.config.Secure.JWT.Cookie.Secure,
		HttpOnly: true, // Securely hide from JS
		SameSite: http.SameSiteLaxMode,
	}

	// Create Refresh Token Cookie
	refreshCookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    loginResponse.RefreshToken,
		Path:     "/",
		Domain:   h.config.Secure.JWT.Cookie.Domain,
		Expires:  loginResponse.RefreshTokenExpiration,
		Secure:   h.config.Secure.JWT.Cookie.Secure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	c.SetCookie(accessCookie)
	c.SetCookie(refreshCookie)

	return c.JSON(http.StatusOK, map[string]string{"message": "success"})
}

func (h *AdminHandler) Dashboard(c echo.Context) error {
	ctx := context.Background()
	// Extract admin info from context (set by JWT middleware)
	adminName := "Admin" // Default fallback
	email, ok := c.Get("email").(string)

	// Try to get user data if we have the user_id or email
	if ok && email != "" {
		// As a fallback or simpler approach, use email as name if name extraction fails or if we want faster load
		adminName = email // Simplification for dashboard display if full profile fetch isn't required

		// Optionally fetch full user profile using UserUseCase if user_id is preferred
		// userID, ok := c.Get("user_id").(string)
		// if ok && userID != "" {
		//     if user, err := h.userUseCase.GetByID(ctx, userID); err == nil {
		//         adminName = user.Name
		//     }
		// }
	}

	// Fetch statistics
	totalUsers := 0
	totalGames := 0
	activeMatches := 0

	// Get total users
	if total, err := h.userUseCase.CountAll(ctx); err == nil {
		totalUsers = total
	}

	// Get total games
	if total, err := h.gameUseCase.CountAll(ctx, nil); err == nil {
		totalGames = total
	}

	// Mock active matches (since MatchUseCase might not have a simple GetAll Active)
	// Optionally replace with a real DB query if added to domain/match
	activeMatches = 42

	adminPath := h.config.App.AdminPath
	if adminPath == "" {
		adminPath = "/admin"
	}

	// Render the admin dashboard template
	component := admin.Dashboard(adminName, totalUsers, totalGames, activeMatches, adminPath)
	return Render(c, http.StatusOK, component)
}

func (h *AdminHandler) Users(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 {
		limit = 10
	}

	ctx := c.Request().Context()
	data, err := h.userUseCase.GetPaginated(ctx, page, limit)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load users")
	}

	adminPath := h.config.App.AdminPath
	if adminPath == "" {
		adminPath = "/admin"
	}
	bucketName := h.config.GCP.BucketName

	if c.Request().Header.Get("HX-Request") == "true" {
		return Render(c, http.StatusOK, admin.UserTableRows(data, adminPath, bucketName))
	}
	return Render(c, http.StatusOK, admin.UsersPage(data, adminPath, bucketName))
}

func (h *AdminHandler) Games(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 {
		limit = 10
	}

	ctx := c.Request().Context()
	data, err := h.gameUseCase.GetPaginated(ctx, page, limit, nil)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to load games")
	}

	adminPath := h.config.App.AdminPath
	if adminPath == "" {
		adminPath = "/admin"
	}
	bucketName := h.config.GCP.BucketName

	if c.Request().Header.Get("HX-Request") == "true" {
		return Render(c, http.StatusOK, admin.GameTableRows(data, adminPath, bucketName))
	}
	return Render(c, http.StatusOK, admin.GamesPage(data, adminPath, bucketName))
}

func (h *AdminHandler) ToggleGameVisibility(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, ErrResponse(domain.ErrInvalidInput))
	}

	ctx := c.Request().Context()
	game, err := h.gameUseCase.GetByID(ctx, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrResponse(domain.ErrNotFound))
	}

	newVisibility := !game.IsPublic
	req := &domain.UpdateGameRequest{
		IsPublic: &newVisibility,
	}

	updatedGame, err := h.gameUseCase.Update(ctx, id, req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrResponse(domain.ErrInternal))
	}

	adminPath := h.config.App.AdminPath
	if adminPath == "" {
		adminPath = "/admin"
	}
	bucketName := h.config.GCP.BucketName

	if c.Request().Header.Get("HX-Request") == "true" {
		// Just return the updated row for HTMX
		return Render(c, http.StatusOK, admin.GameTableRow(*updatedGame, adminPath, bucketName))
	}

	return c.Redirect(http.StatusFound, adminPath+"/games")
}
