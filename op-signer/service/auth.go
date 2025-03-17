package service

import (
	"context"
	"net"
	"net/http"

	oprpc "github.com/ethereum-optimism/optimism/op-service/rpc"
	optls "github.com/ethereum-optimism/optimism/op-service/tls"
)

type ClientInfo struct {
	ClientName string
}

type clientInfoContextKey struct{}

func NewAuthMiddleware() oprpc.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientInfo := ClientInfo{}

			// PeerTLSInfo is attached to context by upstream op-service middleware
			peerTlsInfo := optls.PeerTLSInfoFromContext(r.Context())
			if peerTlsInfo.LeafCertificate == nil {
				http.Error(w, "client certificate was not provided", 401)
				return
			}
			// Note that the certificate is already verified by http server if we get here
			if len(peerTlsInfo.LeafCertificate.IPAddresses) < 1 {
				http.Error(w, "client certificate verified but did not contain IP SAN extension", 401)
				return
			}
			certIP := peerTlsInfo.LeafCertificate.IPAddresses[0].String()
			reqIP, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				http.Error(w, "failed to parse remote address", 400)
				return
			}
			// Check that the IP in the client certificate matches the remote IP address from the request
			if certIP != reqIP {
				http.Error(w, "client certificate IP does not match remote address", 401)
				return
			}
			clientInfo.ClientName = certIP

			ctx := context.WithValue(r.Context(), clientInfoContextKey{}, clientInfo)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ClientInfoFromContext(ctx context.Context) ClientInfo {
	info, _ := ctx.Value(clientInfoContextKey{}).(ClientInfo)
	return info
}
