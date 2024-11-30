package controllers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/LombardiDaniel/go-gin-template/common"
	"github.com/LombardiDaniel/go-gin-template/fiddlers"
	"github.com/LombardiDaniel/go-gin-template/middlewares"
	"github.com/LombardiDaniel/go-gin-template/schemas"
	"github.com/LombardiDaniel/go-gin-template/services"
	"github.com/gin-gonic/gin"
)

type OrganizationController struct {
	userService  services.UserService
	emailService services.EmailService
	orgService   services.OrganizationService
}

func NewOrganizationController(
	userService services.UserService,
	emailService services.EmailService,
	orgService services.OrganizationService,
) OrganizationController {
	return OrganizationController{
		userService:  userService,
		emailService: emailService,
		orgService:   orgService,
	}
}

// @Summary CreateOrganization
// @Security JWT
// @Tags Organization
// @Description Creates an Organization
// @Consume application/json
// @Accept json
// @Produce plain
// @Param   payload 	body 		schemas.CreateOrganization true "org json"
// @Success 200 		{object} 	schemas.IdString
// @Failure 400 		{string} 	ErrorResponse "Bad Request"
// @Failure 409 		{string} 	ErrorResponse "Conflict"
// @Failure 502 		{string} 	ErrorResponse "Bad Gateway"
// @Router /v1/organizations [PUT]
func (c *OrganizationController) CreateOrganization(ctx *gin.Context) {
	rCtx := ctx.Request.Context()
	var createOrg schemas.CreateOrganization

	if err := ctx.ShouldBind(&createOrg); err != nil {
		slog.Error(err.Error())
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	user, err := common.GetClaimsFromGinCtx(ctx)
	if err != nil {
		slog.Error(err.Error())
		ctx.String(http.StatusBadGateway, "BadGateway")
		return
	}

	org, err := fiddlers.NewOrganization(createOrg.OrganizationName, user.UserId)
	if err != nil {
		slog.Error(fmt.Sprintf("Error while generating organization: %s", err.Error()))
		ctx.String(http.StatusBadGateway, "BadGateway")
		return
	}

	err = c.orgService.CreateOrganization(rCtx, *org)
	if err != nil {
		if err == common.ErrDbConflict {
			ctx.String(http.StatusConflict, "Conflict")
			return
		}
		slog.Error(err.Error())
		ctx.String(http.StatusBadGateway, "BadGateway")
		return
	}

	ctx.JSON(http.StatusOK, schemas.IdString{Id: org.OrganizationId})
}

// @Summary InviteToOrg
// @Security JWT
// @Tags Organization
// @Description Invite User to Org
// @Consume application/json
// @Accept json
// @Produce plain
// @Param	orgId 		path string true "Organization Id"
// @Param   payload 	body 		schemas.CreateOrganizationInvite true "invite json"
// @Success 200 		{string} 	OKResponse "OK"
// @Failure 400 		{string} 	ErrorResponse "Bad Request"
// @Failure 409 		{string} 	ErrorResponse "Conflict"
// @Failure 502 		{string} 	ErrorResponse "Bad Gateway"
// @Router /v1/organizations/{orgId}/invite [PUT]
func (c *OrganizationController) InviteToOrg(ctx *gin.Context) {
	rCtx := ctx.Request.Context()
	var createInv schemas.CreateOrganizationInvite

	if err := ctx.ShouldBind(&createInv); err != nil {
		slog.Error(err.Error())
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	currUser, err := common.GetClaimsFromGinCtx(ctx)
	if err != nil {
		slog.Error(err.Error())
		ctx.String(http.StatusBadGateway, "BadGateway")
		return
	}

	otp, err := common.GenerateRandomString(common.OTP_LEN)
	if err != nil {
		slog.Error(err.Error())
		ctx.String(http.StatusBadGateway, "BadGateway")
		return
	}

	user, err := c.userService.GetUser(rCtx, createInv.UserEmail)
	if err != nil {
		slog.Error(err.Error())
		ctx.String(http.StatusBadGateway, "BadGateway")
		return
	}

	org, err := c.orgService.GetOrganization(rCtx, *currUser.OrganizationId)
	if err != nil {
		slog.Error(err.Error())
		ctx.String(http.StatusBadGateway, "BadGateway")
		return
	}

	inv := fiddlers.NewOrganizationInvite(*currUser.OrganizationId, user.UserId, createInv.IsAdmin, otp)
	err = c.orgService.CreateOrganizationInvite(rCtx, inv)
	if err != nil {
		slog.Error(err.Error())
		ctx.String(http.StatusBadGateway, "BadGateway")
		return
	}

	err = c.emailService.SendOrganizationInvite(user.Email, user.FirstName, otp, org.OrganizationName)
	if err != nil {
		slog.Error(err.Error())
		ctx.String(http.StatusBadGateway, "BadGateway")
		return
	}

	ctx.String(http.StatusOK, "OK")
}

// @Summary AcceptOrgInvite
// @Tags Organization
// @Description Accepts the Organization Invite
// @Consume application/json
// @Accept json
// @Produce plain
// @Param   otp 		query 		string true "OneTimePass sent in email"
// @Success 200 		{string} 	OKResponse "OK"
// @Failure 400 		{string} 	ErrorResponse "Bad Request"
// @Failure 409 		{string} 	ErrorResponse "Conflict"
// @Failure 502 		{string} 	ErrorResponse "Bad Gateway"
// @Router /v1/organizations/accept-invite [GET]
func (c *OrganizationController) AcceptOrgInvite(ctx *gin.Context) {
	rCtx := ctx.Request.Context()
	otp := ctx.Query("otp")

	// currUser, err := common.GetClaimsFromGinCtx(ctx)
	// if err != nil {
	// 	slog.Error(err.Error())
	// 	ctx.String(http.StatusBadGateway, "BadGateway")
	// 	return
	// }

	err := c.orgService.ConfirmOrganizationInvite(rCtx, otp)
	if err != nil {
		slog.Error(err.Error())
		ctx.String(http.StatusBadGateway, "BadGateway")
		return
	}

	ctx.String(http.StatusOK, "OK")
}

// @Summary RemoveFromOrg
// @Security JWT
// @Tags Organization
// @Description Removes User from Org
// @Produce plain
// @Param	orgId 		path string true "Organization Id"
// @Param	userId 		path string true "User Id"
// @Success 200 		{string} 	OKResponse "OK"
// @Failure 400 		{string} 	ErrorResponse "Bad Request"
// @Failure 409 		{string} 	ErrorResponse "Conflict"
// @Failure 502 		{string} 	ErrorResponse "Bad Gateway"
// @Router /v1/organizations/{orgId}/users/{userId} [DELETE]
func (c *OrganizationController) RemoveFromOrg(ctx *gin.Context) {
	rCtx := ctx.Request.Context()
	userId, err := strconv.Atoi(ctx.Param("userId"))
	if err != nil {
		ctx.String(http.StatusBadRequest, "BadRequest")
		return
	}

	var createInv schemas.CreateOrganizationInvite

	if err := ctx.ShouldBind(&createInv); err != nil {
		slog.Error(err.Error())
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	currUser, err := common.GetClaimsFromGinCtx(ctx)
	if err != nil {
		slog.Error(err.Error())
		ctx.String(http.StatusBadGateway, "BadGateway")
		return
	}

	err = c.orgService.RemoveUserFromOrg(rCtx, *currUser.OrganizationId, uint32(userId))
	if err != nil {
		slog.Error(err.Error())
		ctx.String(http.StatusBadGateway, "BadGateway")
		return
	}

	ctx.String(http.StatusOK, "OK")
}

// @Summary ChangeOwner
// @Security JWT
// @Tags Organization
// @Description Removes User from Org
// @Produce plain
// @Param	orgId 		path string true "Organization Id"
// @Param   payload 	body 		schemas.Email true "email json"
// @Success 200 		{string} 	OKResponse "OK"
// @Failure 400 		{string} 	ErrorResponse "Bad Request"
// @Failure 409 		{string} 	ErrorResponse "Conflict"
// @Failure 502 		{string} 	ErrorResponse "Bad Gateway"
// @Router /v1/organizations/{orgId}/owner [POST]
func (c *OrganizationController) ChangeOwner(ctx *gin.Context) {
	rCtx := ctx.Request.Context()

	var tgtEmail schemas.Email

	if err := ctx.ShouldBind(&tgtEmail); err != nil {
		slog.Error(err.Error())
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	currUser, err := common.GetClaimsFromGinCtx(ctx)
	if err != nil {
		slog.Error(err.Error())
		ctx.String(http.StatusBadGateway, "BadGateway")
		return
	}

	org, err := c.orgService.GetOrganization(rCtx, *currUser.OrganizationId)
	if err != nil {
		slog.Error(err.Error())
		ctx.String(http.StatusBadGateway, "BadGateway")
		return
	}

	if currUser.UserId != org.OwnerUserId {
		ctx.String(http.StatusUnauthorized, "StatusUnauthorized")
		return
	}

	tgtUser, err := c.userService.GetUser(rCtx, tgtEmail.Email)
	if err != nil {
		slog.Error(err.Error())
		ctx.String(http.StatusBadGateway, "BadGateway")
		return
	}

	err = c.orgService.SetOrganizationOwner(rCtx, *currUser.OrganizationId, tgtUser.UserId)
	if err != nil {
		slog.Error(err.Error())
		ctx.String(http.StatusBadGateway, "BadGateway")
		return
	}

	ctx.String(http.StatusOK, "OK")
}

func (c *OrganizationController) RegisterRoutes(rg *gin.RouterGroup, authMiddleware middlewares.AuthMiddleware) {
	g := rg.Group("/organizations")

	g.PUT("", authMiddleware.AuthorizeUser(), c.CreateOrganization)
	g.PUT("/:orgId/invite", authMiddleware.AuthorizeOrganization(true), c.InviteToOrg)
	g.POST("/:orgId/owner", authMiddleware.AuthorizeOrganization(true), c.ChangeOwner, authMiddleware.Reauthorize())
	g.GET("/accept-invite", c.AcceptOrgInvite)
	g.DELETE("/:orgId/users/:userId", authMiddleware.AuthorizeOrganization(true), c.RemoveFromOrg)
}
