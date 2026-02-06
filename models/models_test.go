package models

import "testing"

func TestUserPasswordAndRole(t *testing.T) {
	user := &User{Username: "u", Email: "e", Role: UserRoleAdmin}
	if err := user.HashPassword("secret123"); err != nil {
		t.Fatalf("hash failed: %v", err)
	}
	if !user.CheckPassword("secret123") {
		t.Fatalf("expected password match")
	}
	if user.CheckPassword("bad") {
		t.Fatalf("expected password mismatch")
	}

	resp := user.ToResponse()
	if resp.Role != "admin" {
		t.Fatalf("expected admin role")
	}
	if !user.IsAdmin() {
		t.Fatalf("expected IsAdmin true")
	}
}

func TestVoteTypeIsValid(t *testing.T) {
	if !VoteLike.IsValid() || !VoteDislike.IsValid() {
		t.Fatalf("expected valid vote types")
	}
	if VoteType("meh").IsValid() {
		t.Fatalf("expected invalid vote type")
	}
}

func TestArticleResponses(t *testing.T) {
	article := &Article{ID: 1, Title: "t", Slug: "s", Author: User{ID: 2, Username: "u"}}
	resp := article.ToResponse()
	if resp.ID != 1 || resp.Author.ID != 2 {
		t.Fatalf("unexpected response")
	}
	summary := article.ToSummaryResponse()
	if summary.Slug != "s" {
		t.Fatalf("unexpected summary")
	}
}

func TestCommentToResponse(t *testing.T) {
	comment := &Comment{ID: 1, Content: "c", ArticleID: "a", UserID: 2, User: User{ID: 2, Username: "u"}}
	resp := comment.ToResponse()
	if resp.User.ID != 2 || resp.Content != "c" {
		t.Fatalf("unexpected comment response")
	}
}

func TestArticleVoteTableName(t *testing.T) {
	if (ArticleVote{}).TableName() != "article_votes" {
		t.Fatalf("unexpected table name")
	}
}
