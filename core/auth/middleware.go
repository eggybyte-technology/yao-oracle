package auth

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// Authenticator validates API keys and returns the associated namespace.
//
// Implementations must be thread-safe as this interface will be called
// concurrently from multiple gRPC handlers.
//
// Example implementation:
//
//	type ConfigBasedAuth struct {
//	    mu sync.RWMutex
//	    namespaces map[string]string // apikey -> namespace
//	}
//
//	func (a *ConfigBasedAuth) ValidateAPIKey(apiKey string) (string, bool) {
//	    a.mu.RLock()
//	    defer a.mu.RUnlock()
//	    ns, ok := a.namespaces[apiKey]
//	    return ns, ok
//	}
type Authenticator interface {
	// ValidateAPIKey checks if the given API key is valid and returns
	// its associated business namespace.
	//
	// Parameters:
	//   - apiKey: The API key from the client request
	//
	// Returns:
	//   - namespace: The business namespace this API key belongs to
	//   - valid: True if the API key is valid, false otherwise
	ValidateAPIKey(apiKey string) (namespace string, valid bool)
}

// UnaryServerInterceptor returns a gRPC unary server interceptor that
// performs API key authentication and namespace resolution.
//
// The interceptor:
//  1. Extracts the API key from "x-api-key" metadata header
//  2. Validates the API key using the provided Authenticator
//  3. Maps the API key to a business namespace
//  4. Injects the namespace into the request context
//  5. Passes control to the actual handler
//
// Health check endpoints are exempt from authentication.
//
// Parameters:
//   - auth: Authenticator implementation for API key validation
//
// Returns:
//   - grpc.UnaryServerInterceptor: The interceptor function
//
// Errors returned:
//   - "missing metadata": No gRPC metadata in request
//   - "missing api key": No "x-api-key" header found
//   - "invalid api key": API key validation failed
//
// Example:
//
//	auth := config.NewAuthenticator()
//	server := grpc.NewServer(
//	    grpc.UnaryInterceptor(auth.UnaryServerInterceptor(auth)),
//	)
func UnaryServerInterceptor(auth Authenticator) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Skip authentication for health checks
		if info.FullMethod == "/grpc.health.v1.Health/Check" {
			return handler(ctx, req)
		}

		// Extract API key from metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, errors.New("missing metadata")
		}

		apiKeys := md.Get("x-api-key")
		if len(apiKeys) == 0 {
			return nil, errors.New("missing api key")
		}

		apiKey := apiKeys[0]
		namespace, valid := auth.ValidateAPIKey(apiKey)
		if !valid {
			return nil, errors.New("invalid api key")
		}

		// Add namespace to context
		ctx = context.WithValue(ctx, "namespace", namespace)

		return handler(ctx, req)
	}
}

// GetNamespaceFromContext extracts the business namespace from the request context.
//
// This should be called in gRPC handlers after the authentication interceptor
// has processed the request. The namespace is used to isolate cache data
// between different clients/tenants.
//
// Parameters:
//   - ctx: Request context (should have been processed by UnaryServerInterceptor)
//
// Returns:
//   - namespace: The business namespace extracted from context
//   - ok: True if namespace was found, false if missing (authentication likely failed)
//
// Example:
//
//	func (s *Server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
//	    namespace, ok := auth.GetNamespaceFromContext(ctx)
//	    if !ok {
//	        return nil, status.Error(codes.Unauthenticated, "namespace not found")
//	    }
//	    // Use namespace for data isolation
//	    value := s.cache.Get(namespace + ":" + req.Key)
//	    return &pb.GetResponse{Value: value}, nil
//	}
func GetNamespaceFromContext(ctx context.Context) (string, bool) {
	namespace, ok := ctx.Value("namespace").(string)
	return namespace, ok
}

// ExtractAPIKeyFromRequest is a helper to extract and validate API key format.
//
// This is useful for services that don't use gRPC interceptors (e.g., HTTP/REST
// endpoints) or for testing purposes.
//
// Parameters:
//   - apiKey: Raw API key string from request
//
// Returns:
//   - string: The API key (currently just returns the input, but provides
//     a hook for future validation/normalization)
//
// Example:
//
//	// In an HTTP handler
//	apiKey := r.Header.Get("X-API-Key")
//	apiKey = auth.ExtractAPIKeyFromRequest(apiKey)
//	namespace, valid := authenticator.ValidateAPIKey(apiKey)
func ExtractAPIKeyFromRequest(apiKey string) string {
	return apiKey
}
