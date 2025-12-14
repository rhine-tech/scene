package mcp

import (
	"context"

	"github.com/rhine-tech/scene"
)

type sceneContextKey struct{}

func createContext(ctx context.Context) scene.Context {
	if v := ctx.Value(sceneContextKey{}); v != nil {
		if sc, ok := v.(scene.Context); ok {
			return sc
		}
	}
	return nil
}

func GetContext(ctx context.Context) (context.Context, scene.Context) {
	if existing := createContext(ctx); existing != nil {
		return ctx, existing
	}
	newCtx := scene.NewContext()
	return context.WithValue(ctx, sceneContextKey{}, newCtx), newCtx
}
