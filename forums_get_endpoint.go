package main

import (
	"strconv"

	"golang.org/x/net/context"
)

var (
	forumGetResponse = map[string]string{
		strconv.FormatInt(ForumIDStarCraft, 10):  ForumNameStarCraft,
		strconv.FormatInt(ForumIDStarCraft2, 10): ForumNameStarCraft2,
		strconv.FormatInt(ForumIDOtherGames, 10): ForumNameOtherGames,
		strconv.FormatInt(ForumIDOffTopic, 10):   ForumNameOffTopic,
	}
)

// ForumsGetEndpoint returns list of all forums.
func ForumsGetEndpoint(ctx context.Context, request interface{}) (interface{}, error) {
	return forumGetResponse, nil
}
