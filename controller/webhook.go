package controller

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http/httputil"
	"net/url"

	"github.com/fabric8-services/fabric8-webhook/app"
	"github.com/fabric8-services/fabric8-webhook/build"
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
	build        build.Service
}

// NewWebhookController creates a Webhook controller.
func NewWebhookController(service *goa.Service,
	config webhookControllerConfiguration,
	vs verification.Service,
	bs build.Service) *WebhookController {
	return &WebhookController{
		Controller:   service.NewController("WebhookController"),
		config:       config,
		verification: vs,
		build:        bs,
	}
}

//GHHookStruct a simplified structure to get info from
//a webhook request
type GHHookStruct struct {
	Repository struct {
		Name     string `json:"name"`
		FullName string `json:"full_name"`
		GitURL   string `json:"git_url"`
		CloneURL string `json:"clone_url"`
	} `json:"repository"`
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

	body, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		return err
	}

	gh := GHHookStruct{}
	err = json.Unmarshal(body, &gh)
	if err != nil {
		return err
	}

	envType, err := c.build.GetEnvironmentType(gh.Repository.GitURL)
	if err != nil {
		return err
	}
	switch envType {
	case "OSIO":
		u, _ := url.Parse(c.config.GetProxyURL())
		proxy := httputil.NewSingleHostReverseProxy(u)
		proxy.ServeHTTP(ctx.ResponseData, ctx.Request)
	case "OSD":
	//TODO
	default:
		return errors.New("Invalid Environment Type")

	}
	return nil
}
