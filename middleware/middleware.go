package middleware

import (
	"gopkg.in/macaron.v1"
)

func SetMiddlewares(m *macaron.Macaron) {
	m.Use(logger())

	m.Use(macaron.Recovery())
}
