package session

import (
	client "github.com/ory/kratos-client-go"
)

// KratosSession is used to access information from a Kratos 'Session'
// JSON payload
type KratosSession struct {
	session *client.Session
}

func NewKratosSession(s client.Session) *client.Session {
	return &s
}
