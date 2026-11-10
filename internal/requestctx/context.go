package httpctx

import "context"

type actor struct {
	userID      int64
	accessToken string
}

type actorKey struct{}

func With(ctx context.Context, userID int64, accessToken string) context.Context {
	return context.WithValue(ctx, actorKey{}, actor{userID: userID, accessToken: accessToken})
}

func UserID(ctx context.Context) (int64, bool) {
	actor, ok := ctx.Value(actorKey{}).(actor)
	if !ok || actor.userID == 0 {
		return 0, false
	}
	return actor.userID, true
}

func AccessToken(ctx context.Context) (string, bool) {
	actor, ok := ctx.Value(actorKey{}).(actor)
	if !ok || actor.accessToken == "" {
		return "", false
	}
	return actor.accessToken, true
}
