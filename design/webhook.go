package design

import (
	d "github.com/goadesign/goa/design"
	a "github.com/goadesign/goa/design/apidsl"
)

var _ = a.Resource("Webhook", func() {

	a.BasePath("/webhook")

	a.Action("forward", func() {
		a.Routing(
			a.POST(""),
		)
		a.Description("Get the current webhook request and forward it" +
			" after verification")
		a.Response(d.OK)
		a.Response(d.Unauthorized)
	})

})
