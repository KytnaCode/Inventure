package web

import "log/slog"

const logKeyRequestID = "request-id"

func logRequestID(reqID string) slog.Attr {
	return slog.String(logKeyRequestID, reqID)
}
