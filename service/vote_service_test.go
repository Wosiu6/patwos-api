package service

import (
	"context"
	"testing"

	"github.com/Wosiu6/patwos-api/models"
)

type voteKey struct {
	articleID uint
	userID    uint
}

type fakeVoteRepo struct {
	items map[voteKey]*models.ArticleVote
}

func newFakeVoteRepo() *fakeVoteRepo {
	return &fakeVoteRepo{items: make(map[voteKey]*models.ArticleVote)}
}

func (r *fakeVoteRepo) Create(_ context.Context, vote *models.ArticleVote) error {
	r.items[voteKey{articleID: vote.ArticleID, userID: vote.UserID}] = vote
	return nil
}

func (r *fakeVoteRepo) Update(_ context.Context, vote *models.ArticleVote) error {
	r.items[voteKey{articleID: vote.ArticleID, userID: vote.UserID}] = vote
	return nil
}

func (r *fakeVoteRepo) Delete(_ context.Context, articleID uint, userID uint) error {
	delete(r.items, voteKey{articleID: articleID, userID: userID})
	return nil
}

func (r *fakeVoteRepo) FindByArticleAndUser(_ context.Context, articleID uint, userID uint) (*models.ArticleVote, error) {
	vote, ok := r.items[voteKey{articleID: articleID, userID: userID}]
	if !ok {
		return nil, nil
	}
	return vote, nil
}

func (r *fakeVoteRepo) CountByArticleAndType(_ context.Context, articleID uint, voteType models.VoteType) (int64, error) {
	var count int64
	for _, v := range r.items {
		if v.ArticleID == articleID && v.VoteType == voteType {
			count++
		}
	}
	return count, nil
}

func (r *fakeVoteRepo) GetVoteCounts(ctx context.Context, articleID uint, userID *uint) (*models.VoteCounts, error) {
	counts := &models.VoteCounts{ArticleID: articleID}

	likes, _ := r.CountByArticleAndType(ctx, articleID, models.VoteLike)
	dislikes, _ := r.CountByArticleAndType(ctx, articleID, models.VoteDislike)
	counts.Likes = likes
	counts.Dislikes = dislikes

	if userID != nil {
		if vote, _ := r.FindByArticleAndUser(ctx, articleID, *userID); vote != nil {
			counts.UserHasVoted = true
			counts.UserVote = string(vote.VoteType)
		}
	}
	return counts, nil
}

func TestVoteService_Flows(t *testing.T) {
	ctx := context.Background()
	repo := newFakeVoteRepo()
	svc := NewVoteService(repo)

	if err := svc.Vote(ctx, 1, 1, "bad"); err != ErrInvalidVoteType {
		t.Fatalf("expected invalid vote type")
	}

	if err := svc.Vote(ctx, 1, 1, models.VoteLike); err != nil {
		t.Fatalf("vote failed: %v", err)
	}

	if err := svc.Vote(ctx, 1, 1, models.VoteDislike); err != nil {
		t.Fatalf("vote update failed: %v", err)
	}

	counts, err := svc.GetVoteCounts(ctx, 1, ptrUint(1))
	if err != nil || counts.Dislikes != 1 || !counts.UserHasVoted {
		t.Fatalf("expected counts with user vote")
	}

	if err := svc.RemoveVote(ctx, 1, 1); err != nil {
		t.Fatalf("remove vote failed: %v", err)
	}
}

func ptrUint(v uint) *uint {
	return &v
}
