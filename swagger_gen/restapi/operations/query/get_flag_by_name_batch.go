// Code generated by go-swagger; DO NOT EDIT.

package query

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"
)

// GetFlagByNameBatchHandlerFunc turns a function with the right signature into a get flag by name batch handler
type GetFlagByNameBatchHandlerFunc func(GetFlagByNameBatchParams) middleware.Responder

// Handle executing the request and returning a response
func (fn GetFlagByNameBatchHandlerFunc) Handle(params GetFlagByNameBatchParams) middleware.Responder {
	return fn(params)
}

// GetFlagByNameBatchHandler interface for that can handle valid get flag by name batch params
type GetFlagByNameBatchHandler interface {
	Handle(GetFlagByNameBatchParams) middleware.Responder
}

// NewGetFlagByNameBatch creates a new http.Handler for the get flag by name batch operation
func NewGetFlagByNameBatch(ctx *middleware.Context, handler GetFlagByNameBatchHandler) *GetFlagByNameBatch {
	return &GetFlagByNameBatch{Context: ctx, Handler: handler}
}

/*GetFlagByNameBatch swagger:route POST /query/batch query getFlagByNameBatch

GetFlagByNameBatch get flag by name batch API

*/
type GetFlagByNameBatch struct {
	Context *middleware.Context
	Handler GetFlagByNameBatchHandler
}

func (o *GetFlagByNameBatch) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewGetFlagByNameBatchParams()

	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request

	o.Context.Respond(rw, r, route.Produces, route, res)

}
