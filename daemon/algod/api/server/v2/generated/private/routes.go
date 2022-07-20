// Package private provides primitives to interact the openapi HTTP API.
//
// Code generated by github.com/algorand/oapi-codegen DO NOT EDIT.
package private

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"github.com/algorand/oapi-codegen/pkg/runtime"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
)

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Aborts a catchpoint catchup.
	// (DELETE /v2/catchup/{catchpoint})
	AbortCatchup(ctx echo.Context, catchpoint string) error
	// Starts a catchpoint catchup.
	// (POST /v2/catchup/{catchpoint})
	StartCatchup(ctx echo.Context, catchpoint string) error
	// Return a list of participation keys
	// (GET /v2/participation)
	GetParticipationKeys(ctx echo.Context) error
	// Add a participation key to the node
	// (POST /v2/participation)
	AddParticipationKey(ctx echo.Context) error
	// Delete a given participation key by ID
	// (DELETE /v2/participation/{participation-id})
	DeleteParticipationKeyByID(ctx echo.Context, participationId string) error
	// Get participation key info given a participation ID
	// (GET /v2/participation/{participation-id})
	GetParticipationKeyByID(ctx echo.Context, participationId string) error
	// Append state proof keys to a participation key
	// (POST /v2/participation/{participation-id})
	AppendKeys(ctx echo.Context, participationId string) error

	// (POST /v2/shutdown)
	ShutdownNode(ctx echo.Context, params ShutdownNodeParams) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// AbortCatchup converts echo context to params.
func (w *ServerInterfaceWrapper) AbortCatchup(ctx echo.Context) error {

	validQueryParams := map[string]bool{
		"pretty": true,
	}

	// Check for unknown query parameters.
	for name, _ := range ctx.QueryParams() {
		if _, ok := validQueryParams[name]; !ok {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Unknown parameter detected: %s", name))
		}
	}

	var err error
	// ------------- Path parameter "catchpoint" -------------
	var catchpoint string

	err = runtime.BindStyledParameter("simple", false, "catchpoint", ctx.Param("catchpoint"), &catchpoint)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter catchpoint: %s", err))
	}

	ctx.Set("api_key.Scopes", []string{""})

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.AbortCatchup(ctx, catchpoint)
	return err
}

// StartCatchup converts echo context to params.
func (w *ServerInterfaceWrapper) StartCatchup(ctx echo.Context) error {

	validQueryParams := map[string]bool{
		"pretty": true,
	}

	// Check for unknown query parameters.
	for name, _ := range ctx.QueryParams() {
		if _, ok := validQueryParams[name]; !ok {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Unknown parameter detected: %s", name))
		}
	}

	var err error
	// ------------- Path parameter "catchpoint" -------------
	var catchpoint string

	err = runtime.BindStyledParameter("simple", false, "catchpoint", ctx.Param("catchpoint"), &catchpoint)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter catchpoint: %s", err))
	}

	ctx.Set("api_key.Scopes", []string{""})

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.StartCatchup(ctx, catchpoint)
	return err
}

// GetParticipationKeys converts echo context to params.
func (w *ServerInterfaceWrapper) GetParticipationKeys(ctx echo.Context) error {

	validQueryParams := map[string]bool{
		"pretty": true,
	}

	// Check for unknown query parameters.
	for name, _ := range ctx.QueryParams() {
		if _, ok := validQueryParams[name]; !ok {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Unknown parameter detected: %s", name))
		}
	}

	var err error

	ctx.Set("api_key.Scopes", []string{""})

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetParticipationKeys(ctx)
	return err
}

// AddParticipationKey converts echo context to params.
func (w *ServerInterfaceWrapper) AddParticipationKey(ctx echo.Context) error {

	validQueryParams := map[string]bool{
		"pretty": true,
	}

	// Check for unknown query parameters.
	for name, _ := range ctx.QueryParams() {
		if _, ok := validQueryParams[name]; !ok {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Unknown parameter detected: %s", name))
		}
	}

	var err error

	ctx.Set("api_key.Scopes", []string{""})

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.AddParticipationKey(ctx)
	return err
}

// DeleteParticipationKeyByID converts echo context to params.
func (w *ServerInterfaceWrapper) DeleteParticipationKeyByID(ctx echo.Context) error {

	validQueryParams := map[string]bool{
		"pretty": true,
	}

	// Check for unknown query parameters.
	for name, _ := range ctx.QueryParams() {
		if _, ok := validQueryParams[name]; !ok {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Unknown parameter detected: %s", name))
		}
	}

	var err error
	// ------------- Path parameter "participation-id" -------------
	var participationId string

	err = runtime.BindStyledParameter("simple", false, "participation-id", ctx.Param("participation-id"), &participationId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter participation-id: %s", err))
	}

	ctx.Set("api_key.Scopes", []string{""})

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.DeleteParticipationKeyByID(ctx, participationId)
	return err
}

// GetParticipationKeyByID converts echo context to params.
func (w *ServerInterfaceWrapper) GetParticipationKeyByID(ctx echo.Context) error {

	validQueryParams := map[string]bool{
		"pretty": true,
	}

	// Check for unknown query parameters.
	for name, _ := range ctx.QueryParams() {
		if _, ok := validQueryParams[name]; !ok {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Unknown parameter detected: %s", name))
		}
	}

	var err error
	// ------------- Path parameter "participation-id" -------------
	var participationId string

	err = runtime.BindStyledParameter("simple", false, "participation-id", ctx.Param("participation-id"), &participationId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter participation-id: %s", err))
	}

	ctx.Set("api_key.Scopes", []string{""})

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetParticipationKeyByID(ctx, participationId)
	return err
}

// AppendKeys converts echo context to params.
func (w *ServerInterfaceWrapper) AppendKeys(ctx echo.Context) error {

	validQueryParams := map[string]bool{
		"pretty": true,
	}

	// Check for unknown query parameters.
	for name, _ := range ctx.QueryParams() {
		if _, ok := validQueryParams[name]; !ok {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Unknown parameter detected: %s", name))
		}
	}

	var err error
	// ------------- Path parameter "participation-id" -------------
	var participationId string

	err = runtime.BindStyledParameter("simple", false, "participation-id", ctx.Param("participation-id"), &participationId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter participation-id: %s", err))
	}

	ctx.Set("api_key.Scopes", []string{""})

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.AppendKeys(ctx, participationId)
	return err
}

// ShutdownNode converts echo context to params.
func (w *ServerInterfaceWrapper) ShutdownNode(ctx echo.Context) error {

	validQueryParams := map[string]bool{
		"pretty":  true,
		"timeout": true,
	}

	// Check for unknown query parameters.
	for name, _ := range ctx.QueryParams() {
		if _, ok := validQueryParams[name]; !ok {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Unknown parameter detected: %s", name))
		}
	}

	var err error

	ctx.Set("api_key.Scopes", []string{""})

	// Parameter object where we will unmarshal all parameters from the context
	var params ShutdownNodeParams
	// ------------- Optional query parameter "timeout" -------------
	if paramValue := ctx.QueryParam("timeout"); paramValue != "" {

	}

	err = runtime.BindQueryParameter("form", true, false, "timeout", ctx.QueryParams(), &params.Timeout)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter timeout: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.ShutdownNode(ctx, params)
	return err
}

// RegisterHandlers adds each server route to the EchoRouter.
func RegisterHandlers(router interface {
	CONNECT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	TRACE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}, si ServerInterface, m ...echo.MiddlewareFunc) {

	wrapper := ServerInterfaceWrapper{
		Handler: si,
	}

	router.DELETE("/v2/catchup/:catchpoint", wrapper.AbortCatchup, m...)
	router.POST("/v2/catchup/:catchpoint", wrapper.StartCatchup, m...)
	router.GET("/v2/participation", wrapper.GetParticipationKeys, m...)
	router.POST("/v2/participation", wrapper.AddParticipationKey, m...)
	router.DELETE("/v2/participation/:participation-id", wrapper.DeleteParticipationKeyByID, m...)
	router.GET("/v2/participation/:participation-id", wrapper.GetParticipationKeyByID, m...)
	router.POST("/v2/participation/:participation-id", wrapper.AppendKeys, m...)
	router.POST("/v2/shutdown", wrapper.ShutdownNode, m...)

}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"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",
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file.
func GetSwagger() (*openapi3.Swagger, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %s", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}

	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("error loading Swagger: %s", err)
	}
	return swagger, nil
}
