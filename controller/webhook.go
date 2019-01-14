package controller

import (
	"errors"
	"net/http/httputil"
	"net/url"

	"github.com/fabric8-services/fabric8-webhook/app"
	"github.com/fabric8-services/fabric8-webhook/verification"
	"github.com/goadesign/goa"
)

// WebhookControllerConfiguration the Configuration for the WebhookController
type webhookControllerConfiguration interface {
	GetProxyURL() string
}

// WebhookController implements the Webhook resource.
type WebhookController struct {
	*goa.Controller
	config       webhookControllerConfiguration
	verification verification.Service
}

// NewWebhookController creates a Webhook controller.
func NewWebhookController(service *goa.Service,
	config webhookControllerConfiguration,
	vs verification.Service) *WebhookController {
	return &WebhookController{
		Controller: service.NewController("WebhookController"),
		config:     config}
}

// Forward runs the forward action.
func (c *WebhookController) Forward(ctx *app.ForwardWebhookContext) error {
	// WebhookController_Forward: start_implement

	// Put your logic here

	// WebhookController_Forward: end_implement
	isVerify, err := c.verification.Verify(ctx.Request)
	if err != nil {
		c.Service.LogInfo("Error while verifying", "err:", err)
		return err
	}
	if !isVerify {
		return errors.New("Request from unauthorized source")
	}

	u, _ := url.Parse(c.config.GetProxyURL())
	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.ServeHTTP(ctx.ResponseData, ctx.Request)
	return nil
}
