// Package auth provides API key authentication and authorization for
// the Yao-Oracle distributed cache system.
//
// This package implements middleware for gRPC services to validate API keys
// and map them to business namespaces, providing multi-tenancy support.
//
// # Basic Usage
//
// Implement the Authenticator interface to provide API key validation:
//
//	type MyAuth struct {
//	    apiKeys map[string]string // apikey -> namespace
//	}
//
//	func (a *MyAuth) ValidateAPIKey(apiKey string) (namespace string, valid bool) {
//	    ns, ok := a.apiKeys[apiKey]
//	    return ns, ok
//	}
//
// Use the interceptor in your gRPC server:
//
//	auth := &MyAuth{...}
//	server := grpc.NewServer(
//	    grpc.UnaryInterceptor(auth.UnaryServerInterceptor(auth)),
//	)
//
// Extract namespace from context in your handlers:
//
//	func (s *Server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
//	    namespace, ok := auth.GetNamespaceFromContext(ctx)
//	    if !ok {
//	        return nil, errors.New("namespace not found in context")
//	    }
//	    // Use namespace to isolate data
//	    value := s.cache.Get(namespace, req.Key)
//	    return &pb.GetResponse{Value: value}, nil
//	}
//
// # Namespace Isolation
//
// The authentication middleware automatically:
//  1. Extracts API key from "x-api-key" metadata header
//  2. Validates the API key
//  3. Maps it to a business namespace
//  4. Injects the namespace into the request context
//
// This enables multi-tenancy where different clients (identified by API keys)
// have isolated namespaces for their cache data.
package auth
