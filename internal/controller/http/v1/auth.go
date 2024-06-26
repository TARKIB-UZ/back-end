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
		h.POST("/forgot", r.forgotPassword)
		h.POST("/reset", r.resetPassword)
		h.POST("/login", r.login)
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
		if err.Error() == "this nickname is already taken" {
			r.l.Error(err, "http - v1 - register")
			errorResponse(c, http.StatusBadRequest, "Sorry, this nickname is already taken")
		} else if err.Error() == "user with this phone number already registered" {
			r.l.Error(err, "http - v1 - register")
			errorResponse(c, http.StatusBadRequest, "Sorry, user with this phone number is already registered")
		} else {
			r.l.Error(err, "http - v1 - register")
			errorResponse(c, http.StatusInternalServerError, "auth service problems")
		}
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
		if err.Error() == "verification code expired" {
			r.l.Error("http - v1 - register")
			errorResponse(c, http.StatusBadRequest, "Verification code expired.")
			return
		} else if err.Error() == "redis: nil" {
			r.l.Error("http - v1 - register")
			errorResponse(c, http.StatusBadRequest, "You have entered wrong phone number.")
			return
		} else if err.Error() == "invalid verification code" {
			r.l.Error("http - v1 - register")
			errorResponse(c, http.StatusBadRequest, "Invalid verification code.")
			return
		} else {
			r.l.Error(err, "http - v1 - register")
			errorResponse(c, http.StatusInternalServerError, "auth service problems")
			return
		}
	}

	c.JSON(http.StatusOK, user)
}

// @Summary     Forgot Password
// @Description Initiates the password reset process by sending a reset code to the user's phone number.
// @ID          forgot-password
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       request body models.ForgotPasswordRequest true "Phone number"
// @Success     200 {object} response
// @Failure     400 {object} response
// @Failure     500 {object} response
// @Router      /auth/forgot [post]
func (r *authRoutes) forgotPassword(c *gin.Context) {
	var request models.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		r.l.Error(err, "http - v1 - forgotPassword")
		errorResponse(c, http.StatusBadRequest, "invalid request body")
		return
	}

	err := r.t.ForgotPassword(c.Request.Context(), request.PhoneNumber)
	if err != nil {
		if err.Error() == "this phone number not registered in tarkib.uz yet" {
			r.l.Error(err, "http - v1 - forgotPassword")
			errorResponse(c, http.StatusBadRequest, "This phone number is not registered in tarkib.uz yet")
			return
		} else {
			r.l.Error(err, "http - v1 - forgotPassword")
			errorResponse(c, http.StatusInternalServerError, "auth service problems")
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset code sent to your phone number.",
	})
}

// @Summary     Reset Password
// @Description Resets the user's password using the provided reset code and new password.
// @ID          reset-password
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       request body models.ResetPasswordRequest true "Phone number, reset code, and new password"
// @Success     200 {object} models.ResetPasswordResponse
// @Failure     400 {object} response
// @Failure     500 {object} response
// @Router      /auth/reset [post]
func (r *authRoutes) resetPassword(c *gin.Context) {
	var request models.ResetPasswordRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		r.l.Error(err, "http - v1 - resetPassword")
		errorResponse(c, http.StatusBadRequest, "invalid request body")
		return
	}

	err := r.t.ResetPassword(
		c.Request.Context(),
		request.PhoneNumber,
		request.Code,
		request.NewPassword,
	)
	if err != nil {
		if err.Error() == "invalid reset code" {
			r.l.Error(err, "http - v1 - resetPassword")
			errorResponse(c, http.StatusBadRequest, "You have entered wrong code.")
		}
		r.l.Error(err, "http - v1 - resetPassword")
		errorResponse(c, http.StatusInternalServerError, "auth service problems")
		return
	}

	c.JSON(http.StatusOK, models.ResetPasswordResponse{
		Message: "Password reset successfully.",
	})
}

// @Summary     Login
// @Description Authenticates a user and returns an access token on successful login.
// @ID          login-user
// @Tags  	    auth
// @Accept      json
// @Produce     json
// @Param       request body models.LoginRequest true "Nickname or Phone Number and Password"
// @Success     200 {object} models.LoginResponse
// @Failure     400 {object} response
// @Failure     401 {object} response
// @Failure     500 {object} response
// @Router      /auth/login [post]
func (r *authRoutes) login(c *gin.Context) {
	var request models.LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		r.l.Error(err, "http - v1 - login")
		errorResponse(c, http.StatusBadRequest, "invalid request body")
		return
	}

	response, err := r.t.Login(c.Request.Context(), entity.LoginRequest{
		NickName:    request.NickName,
		PhoneNumber: request.PhoneNumber,
		Password:    request.Password,
	})
	if err != nil {
		switch err.Error() {
		case "user not found":
			r.l.Error(err, "http - v1 - login")
			errorResponse(c, http.StatusUnauthorized, "Invalid nickname or phone number")
		case "invalid password":
			r.l.Error(err, "http - v1 - login")
			errorResponse(c, http.StatusUnauthorized, "Invalid password")
		default:
			r.l.Error(err, "http - v1 - login")
			errorResponse(c, http.StatusInternalServerError, "auth service problems")
		}
		return
	}

	c.JSON(http.StatusOK, response)
}
