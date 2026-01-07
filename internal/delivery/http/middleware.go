package httpdelivery

import (
	"strings"

	"github.com/example/learngo/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	CtxUserID = "userId"
	CtxRole   = "role"
)

// AuthRequired валидирует Bearer-токен и кладёт userId/role в контекст.
func AuthRequired(jwt *utils.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			UnauthorizedError(c, "Missing bearer token")
			c.Abort()
			return
		}
		claims, err := jwt.Verify(parts[1])
		if err != nil {
			UnauthorizedError(c, "Invalid token")
			c.Abort()
			return
		}
		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxRole, claims.Role)
		c.Next()
	}
}

// UserIDFromContext helper.
func UserIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	v, ok := c.Get(CtxUserID)
	if !ok {
		return uuid.Nil, false
	}
	id, _ := v.(uuid.UUID)
	return id, id != uuid.Nil
}

// RequireRoles разрешает доступ только пользователям с одной из ролей.
func RequireRoles(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(c *gin.Context) {
		role := c.GetString(CtxRole)
		if _, ok := allowed[role]; !ok {
			ForbiddenError(c, "")
			c.Abort()
			return
		}
		c.Next()
	}
}
