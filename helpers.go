package jrpc2server

import (
	"errors"

	"github.com/pquerna/ffjson/ffjson"
	"github.com/riftbit/jrpc2errors"
	"github.com/valyala/fasthttp"
)

// ReadRequestParams getting request parametrs
func ReadRequestParams(request *ServerRequest, args interface{}) error {
	if request.Params != nil {
		// Note: if c.request.Params is nil it's not an error, it's an optional member.
		// JSON params structured object. Unmarshal to the args object.
		if err := ffjson.Unmarshal(*request.Params, args); err != nil {
			// Clearly JSON params is not a structured object,
			// fallback and attempt an unmarshal with JSON params as
			// array value and RPC params is struct. Unmarshal into
			// array containing the request struct.
			params := [1]interface{}{args}
			if err = ffjson.Unmarshal(*request.Params, &params); err != nil {
				return err
			}
		}
	}
	return nil
}

// WriteResponse write response to client with status code and server response struct
func WriteResponse(ctx *fasthttp.RequestCtx, status int, resp *ServerResponse) {
	body, _ := ffjson.Marshal(resp)
	ctx.SetBody(body)
	ffjson.Pool(body)
	ctx.Response.Header.Set("x-content-type-options", "nosniff")
	ctx.SetContentType("application/json; charset=utf-8")
	ctx.SetStatusCode(status)
}

// PrepareDataHandler process basic data to context values PrepareDataHandlerRequestErr.(error) and PrepareDataHandlerRequest.(*ServerRequest) and PrepareDataHandlerRequestRun.(int)
func PrepareDataHandler(ctx *fasthttp.RequestCtx) {

	var err error

	ctx.SetUserValue("PrepareDataHandlerRequestErr", err)
	ctx.SetUserValue("PrepareDataHandlerRequestRun", 1)

	if string(ctx.Method()) != "POST" {

		err = &jrpc2errors.Error{
			Code:    jrpc2errors.ParseError,
			Message: errors.New("api: POST method required, received " + string(ctx.Method())).Error(),
		}

		resp := &ServerResponse{
			Version: Version,
			Error:   err.(*jrpc2errors.Error),
		}

		ctx.SetUserValue("PrepareDataHandlerRequestErr", err)
		WriteResponse(ctx, 405, resp)
		return
	}

	req := new(ServerRequest)

	err = ffjson.Unmarshal(ctx.Request.Body(), req)
	if err != nil {
		err = &jrpc2errors.Error{
			Code:    jrpc2errors.ParseError,
			Message: err.Error(),
			Data:    req,
		}

		resp := &ServerResponse{
			Version: Version,
			ID:      req.ID,
			Error:   err.(*jrpc2errors.Error),
		}

		ctx.SetUserValue("PrepareDataHandlerRequestErr", err)
		WriteResponse(ctx, 400, resp)
		return
	}

	ctx.SetUserValue("PrepareDataHandlerRequest", req)

	if req.Version != Version {
		err = &jrpc2errors.Error{
			Code:    jrpc2errors.InvalidRequestError,
			Message: "jsonrpc must be " + Version,
			Data:    req,
		}

		resp := &ServerResponse{
			Version: Version,
			ID:      req.ID,
			Error:   err.(*jrpc2errors.Error),
		}

		ctx.SetUserValue("PrepareDataHandlerRequestErr", err)
		WriteResponse(ctx, 400, resp)
		return
	}

}
