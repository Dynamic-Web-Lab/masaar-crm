package ai

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/dynamicweblab/masaar-crm/internal/domain"
	"github.com/dynamicweblab/masaar-crm/internal/repo"
)

type ScoringService struct {
	leadRepo *repo.LeadRepo
	commRepo *repo.CommunicationHistoryRepo
	tagRepo  *repo.LeadTagRepo
}

func NewScoringService(leadRepo *repo.LeadRepo, commRepo *repo.CommunicationHistoryRepo, tagRepo *repo.LeadTagRepo) *ScoringService {
	return &ScoringService{
		leadRepo: leadRepo,
		commRepo: commRepo,
		tagRepo:  tagRepo,
	}
}

// CalculateScore computes a lead score (0-100) based on multiple factors
func (s *ScoringService) CalculateScore(ctx context.Context, leadID uuid.UUID) (int, error) {
	lead, err := s.leadRepo.GetByID(ctx, leadID)
	if err != nil {
		return 0, err
	}

	score := 0

	// Stage progression (0-40 points)
	score += s.scoreByStage(lead.Stage)

	// Recency (0-20 points) — penalty for age
	score += s.scoreByRecency(lead.CreatedAt)

	// Engagement (0-30 points) — based on message frequency and responsiveness
	engScore, err := s.scoreByEngagement(ctx, leadID)
	if err != nil {
		engScore = 0
	}
	score += engScore

	// Quality tags (0-10 points) — hot/warm tag boost
	qualityScore, err := s.scoreByQualityTags(ctx, leadID)
	if err != nil {
		qualityScore = 0
	}
	score += qualityScore

	if score > 100 {
		score = 100
	}
	return score, nil
}

// scoreByStage assigns points based on lead stage progression
func (s *ScoringService) scoreByStage(stage domain.LeadStage) int {
	switch stage {
	case domain.StageNew:
		return 0 // Just entered
	case domain.StageContacted:
		return 10 // We engaged
	case domain.StageQualified:
		return 25 // Qualified as real lead
	case domain.StageProposal:
		return 40 // Serious intent
	case domain.StageWon:
		return 100 // Converted
	case domain.StageLost:
		return 0 // No longer viable
	default:
		return 0
	}
}

// scoreByRecency penalizes old leads but rewards recent activity
func (s *ScoringService) scoreByRecency(createdAt time.Time) int {
	daysSinceCreation := time.Since(createdAt).Hours() / 24

	// Decay function: newer leads score higher
	// 0 days = 20 points
	// 7 days = 15 points
	// 14 days = 10 points
	// 30+ days = 5 points
	if daysSinceCreation <= 1 {
		return 20
	} else if daysSinceCreation <= 7 {
		return 15
	} else if daysSinceCreation <= 14 {
		return 10
	} else if daysSinceCreation <= 30 {
		return 7
	} else {
		return max(1, 5-int(daysSinceCreation/30)) // Floor at 1
	}
}

// scoreByEngagement rewards message frequency and responsiveness
func (s *ScoringService) scoreByEngagement(ctx context.Context, leadID uuid.UUID) (int, error) {
	comms, err := s.commRepo.GetByLead(ctx, leadID, 100)
	if err != nil {
		return 0, err
	}

	if len(comms) == 0 {
		return 0, nil
	}

	score := 0

	// Base engagement: 1+ messages = 5 points
	if len(comms) > 0 {
		score += 5
	}

	// Multiple messages = higher score (max +15)
	score += min(15, len(comms)*2)

	// Recent messages boost (last message in last 24h = +10)
	if len(comms) > 0 && time.Since(comms[0].CreatedAt) < 24*time.Hour {
		score += 10
	}

	return score, nil
}

// scoreByQualityTags applies boost for high-quality indicators
func (s *ScoringService) scoreByQualityTags(ctx context.Context, leadID uuid.UUID) (int, error) {
	qualityTags, err := s.tagRepo.ListByCategory(ctx, leadID, "quality")
	if err != nil {
		return 0, err
	}

	score := 0
	for _, tag := range qualityTags {
		if tag == "hot" {
			score += 10
		} else if tag == "warm" {
			score += 5
		}
	}

	return score, nil
}

// UpdateScoreOnMessage updates lead score after message received
func (s *ScoringService) UpdateScoreOnMessage(ctx context.Context, leadID uuid.UUID) error {
	score, err := s.CalculateScore(ctx, leadID)
	if err != nil {
		return err
	}

	return s.leadRepo.UpdateScore(ctx, leadID, score)
}

// UpdateScoreOnStageChange updates lead score when stage changes
func (s *ScoringService) UpdateScoreOnStageChange(ctx context.Context, leadID uuid.UUID, newStage domain.LeadStage) error {
	score, err := s.CalculateScore(ctx, leadID)
	if err != nil {
		return err
	}

	// Boost for stage progression (qualified and above)
	if newStage == domain.StageQualified || newStage == domain.StageProposal {
		score = min(100, score+5)
	}

	return s.leadRepo.UpdateScore(ctx, leadID, score)
}

// ApplyTimeDecay applies age-based decay to all active leads
// Should be called periodically (e.g., daily via scheduled job)
func (s *ScoringService) ApplyTimeDecay(ctx context.Context) error {
	// For now, this is a stub that would be called by a background job
	// In production, query all leads and recalculate scores
	// This prevents leads from staying hot indefinitely due to age

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
