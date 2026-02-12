package middleware

import (
	"fmt"
	"net/http"

	"github.com/everyday-studio/ollm/internal/domain"
	"github.com/labstack/echo/v4"
)

func AllowRoles(allowedRole domain.Role) echo.MiddlewareFunc {
	rolePriority := map[domain.Role]int{
		domain.RoleAdmin:   3,
		domain.RoleManager: 2,
		domain.RoleUser:    1,
		domain.RolePublic:  0,
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if allowedRole == domain.RolePublic {
				return next(c)
			}

			// Get role from context (set by JWT middleware)
			roleInToken, ok := c.Get("role").(string)
			if !ok || roleInToken == "" {
				return echo.NewHTTPError(http.StatusForbidden, "authentication required")
			}

			if rolePriority[domain.Role(roleInToken)] >= rolePriority[allowedRole] {
				return next(c)
			}

			errMessage := fmt.Sprintf("insufficient permission to access this resource. allowed: %v, user role %v", allowedRole, roleInToken)
			return echo.NewHTTPError(http.StatusForbidden, errMessage)
		}
	}
}
