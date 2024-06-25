package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"tarkib.uz/internal/controller/http/models"
	"tarkib.uz/internal/entity"
	"tarkib.uz/internal/usecase"
	"tarkib.uz/pkg/logger"
)

type authRoutes struct {
	t usecase.Auth
	l logger.Interface
}

func newAuthRoutes(handler *gin.RouterGroup, t usecase.Auth, l logger.Interface) {
	r := &authRoutes{t, l}

	h := handler.Group("/auth")
	{
		h.POST("/register", r.register)
		h.POST("/verify", r.verify)
	}
}

// @Summary     Register
// @Description Registers a new user
// @ID          register-user
// @Tags  	    auth
// @Accept      json
// @Produce     json
// @Param       request body models.RegisterUser true "User credentials"
// @Success     200 {object} models.RegisterUser
// @Failure     400 {object} response
// @Failure     500 {object} response
// @Router      /auth/register [post]
func (r *authRoutes) register(c *gin.Context) {
	var request models.RegisterUser
	if err := c.ShouldBindJSON(&request); err != nil {
		r.l.Error(err, "http - v1 - register")
		errorResponse(c, http.StatusBadRequest, "invalid request body")

		return
	}

	err := r.t.Register(
		c.Request.Context(),
		&entity.User{
			FirstName:   request.FirstName,
			LastName:    request.LastName,
			PhoneNumber: request.PhoneNumber,
			NickName:    request.NickName,
			Password:    request.Password,
			Avatar:      request.Avatar,
		},
	)
	if err != nil {
		r.l.Error(err, "http - v1 - register")
		errorResponse(c, http.StatusInternalServerError, "auth service problems")

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Code sent to your phone number. Please verify.",
	})
}

// @Summary     Verify
// @Description After register, user must be verified.
// @ID          verify-user
// @Tags  	    auth
// @Accept      json
// @Produce     json
// @Param       request body models.VerifyUser true "One time code and phone number"
// @Success     200 {object} models.VerifyUserResponse
// @Failure     400 {object} response
// @Failure     500 {object} response
// @Router      /auth/verify [post]
func (r *authRoutes) verify(c *gin.Context) {
	var request models.VerifyUser
	if err := c.ShouldBindJSON(&request); err != nil {
		r.l.Error(err, "http - v1 - verify")
		errorResponse(c, http.StatusBadRequest, "invalid request body")

		return
	}

	user, err := r.t.Verify(c.Request.Context(), entity.VerifyUser{
		PhoneNumber: request.PhoneNumber,
		Code:        request.Code,
	})
	if err != nil {
		r.l.Error(err, "http - v1 - register")
		errorResponse(c, http.StatusInternalServerError, "auth service problems")

		return
	}

	c.JSON(http.StatusOK, user)
}
