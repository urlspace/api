package handlers

import "time"

const TokenExpiryDuration = 24 * time.Hour
const SessionExpiryDuration = 30 * 24 * time.Hour
const SessionCookieName = "session_id"
const TokenTypeSession = "session"
const TokenTypeManaged = "token"
