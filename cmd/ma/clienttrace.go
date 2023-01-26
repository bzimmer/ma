package main

import (
	"crypto/tls"
	"net/http/httptrace"
	"net/textproto"

	"github.com/rs/zerolog/log"
)

func ClientTrace() *httptrace.ClientTrace {

	return &httptrace.ClientTrace{
		// GetConn is called before a connection is created or
		// retrieved from an idle pool. The hostPort is the
		// "host:port" of the target or proxy. GetConn is called even
		// if there's already an idle cached connection available.
		GetConn: func(hostPort string) {
			log.Info().Str("hostport", hostPort).Msg("GetConn")
		},

		// GotConn is called after a successful connection is
		// obtained. There is no hook for failure to obtain a
		// connection; instead, use the error from
		// Transport.RoundTrip.
		GotConn: func(info httptrace.GotConnInfo) {
			log.Info().Interface("info", info).Msg("GotConn")

		},

		// PutIdleConn is called when the connection is returned to
		// the idle pool. If err is nil, the connection was
		// successfully returned to the idle pool. If err is non-nil,
		// it describes why not. PutIdleConn is not called if
		// connection reuse is disabled via Transport.DisableKeepAlives.
		// PutIdleConn is called before the caller's Response.Body.Close
		// call returns.
		// For HTTP/2, this hook is not currently used.
		PutIdleConn: func(err error) {
			log.Info().Err(err).Msg("PutIdleConn")

		},

		// GotFirstResponseByte is called when the first byte of the response
		// headers is available.
		GotFirstResponseByte: func() {
			log.Info().Msg("GotFirstResponseByte")

		},

		// Got100Continue is called if the server replies with a "100
		// Continue" response.
		Got100Continue: func() {
			log.Info().Msg("Got100Continue")

		},

		// Got1xxResponse is called for each 1xx informational response header
		// returned before the final non-1xx response. Got1xxResponse is called
		// for "100 Continue" responses, even if Got100Continue is also defined.
		// If it returns an error, the client request is aborted with that error value.
		Got1xxResponse: func(code int, header textproto.MIMEHeader) error {
			log.Info().Int("code", code).Msg("Got1xxResponse")

			return nil
		},

		// DNSStart is called when a DNS lookup begins.
		DNSStart: func(info httptrace.DNSStartInfo) {
			log.Info().Interface("info", info).Msg("DNSStart")

		},

		// DNSDone is called when a DNS lookup ends.
		DNSDone: func(info httptrace.DNSDoneInfo) {
			log.Info().Interface("info", info).Msg("DNSDone")
		},

		// ConnectStart is called when a new connection's Dial begins.
		// If net.Dialer.DualStack (IPv6 "Happy Eyeballs") support is
		// enabled, this may be called multiple times.
		ConnectStart: func(network, addr string) {
			log.Info().Str("network", network).Str("addr", addr).Msg("ConnectStart")
		},

		// ConnectDone is called when a new connection's Dial
		// completes. The provided err indicates whether the
		// connection completed successfully.
		// If net.Dialer.DualStack ("Happy Eyeballs") support is
		// enabled, this may be called multiple times.
		ConnectDone: func(network, addr string, err error) {
			log.Info().Str("network", network).Str("addr", addr).Msg("ConnectDone")
		},

		// TLSHandshakeStart is called when the TLS handshake is started. When
		// connecting to an HTTPS site via an HTTP proxy, the handshake happens
		// after the CONNECT request is processed by the proxy.
		TLSHandshakeStart: func() {
			log.Info().Msg("TLSHandshakeStart")
		},

		// TLSHandshakeDone is called after the TLS handshake with either the
		// successful handshake's connection state, or a non-nil error on handshake
		// failure.
		TLSHandshakeDone: func(_ tls.ConnectionState, _ error) {
			log.Info().Msg("TLSHandshakeDone")
		},

		// WroteHeaderField is called after the Transport has written
		// each request header. At the time of this call the values
		// might be buffered and not yet written to the network.
		WroteHeaderField: func(key string, value []string) {
			log.Info().Str("key", key).Strs("value", value).Msg("WroteHeaderField")
		},

		// WroteHeaders is called after the Transport has written
		// all request headers.
		WroteHeaders: func() {
			log.Info().Msg("WroteHeaders")
		},

		// Wait100Continue is called if the Request specified
		// "Expect: 100-continue" and the Transport has written the
		// request headers but is waiting for "100 Continue" from the
		// server before writing the request body.
		Wait100Continue: func() {
			log.Info().Msg("Wait100Continue")
		},

		// WroteRequest is called with the result of writing the
		// request and any body. It may be called multiple times
		// in the case of retried requests.
		WroteRequest: func(info httptrace.WroteRequestInfo) {
			log.Info().Msg("WroteRequest")
		},
	}
}
