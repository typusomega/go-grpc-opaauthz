/*
 *
 * Copyright 2019 Jens Bieber
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package opaauthz

import (
	"context"
	"errors"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
	"gopkg.in/resty.v1"
)

// A AuthzOption sets options such as url, used headers etc.
type AuthzOption func(*OpaAuthorizer)

// OpaURL points to the OPA HTTP Server used as authorization backend.
// example: "http://localhost:8181/v1/data/apis/invocation_allowed"
// Find further information about OPA API [here](https://www.openpolicyagent.org/docs/rest-api.html)
func OpaURL(url string) AuthzOption {
	return func(o *OpaAuthorizer) {
		o.OpaURL = strings.ToLower(url)
	}
}

// CredentialHeader is the name used to extract client credentials.
// default: "authorization"
func CredentialHeader(headerName string) AuthzOption {
	return func(o *OpaAuthorizer) {
		o.CredentialHeader = headerName
	}
}

// NewOpaAuthorizer creates an OPA authorizer
func NewOpaAuthorizer(options ...AuthzOption) *OpaAuthorizer {
	authz := &OpaAuthorizer{
		OpaURL:           "http://localhost:8181/v1/data/apis/invocation_allowed",
		CredentialHeader: "authorization",
	}
	for _, opt := range options {
		opt(authz)
	}
	return authz
}

// OpaAuthorizer is a gRPC server authorizer using OPA as backend
type OpaAuthorizer struct {
	OpaURL           string
	CredentialHeader string
}

// OpaStreamInterceptor is OpaAuthorizers StreamServerInterceptor for the
// server. Only one stream interceptor can be installed.
// If you want to add extra functionality you might decorate this function.
func (authz *OpaAuthorizer) OpaStreamInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if err := authz.authorize(stream.Context(), info.FullMethod); err != nil {
		return err
	}

	return handler(srv, stream)
}

// OpaUnaryInterceptor is OpaAuthorizers UnaryServerInterceptor for the
// server. Only one unary interceptor can be installed.
// If you want to add extra functionality you might decorate this function.
func (authz *OpaAuthorizer) OpaUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if err := authz.authorize(ctx, info.FullMethod); err != nil {
		return nil, err
	}

	return handler(ctx, req)
}

func (authz *OpaAuthorizer) authorize(ctx context.Context, methodName string) error {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if token, exists := md[authz.CredentialHeader]; exists {
			// request opa
			resp, err := resty.R().
				SetBody(newOpaRequest(methodName, token[0])).
				SetResult(&opaSuccessResponse{}).
				Post(authz.OpaURL)
			if err != nil || resp.IsError() {
				grpclog.Errorf("opaauthz: could not request OPA: %v", err.Error())
			}

			response := resp.Result().(*opaSuccessResponse)
			if response.Allowed {
				return nil
			}
		}

		return errors.New("unauthorized")
	}

	return errors.New("empty metadata")
}

type opaSuccessResponse struct {
	Allowed bool `json:"result"`
}

type opaRequest struct {
	Input map[string]interface{} `json:"input"`
}

func newOpaRequest(method string, token string) *opaRequest {
	input := map[string]interface{}{
		"method":    method,
		"authToken": token,
	}
	return &opaRequest{Input: input}
}
