package service

import (
	"context"

	"the-line/backend/internal/domain"
	"the-line/backend/internal/dto"
	"the-line/backend/internal/model"
	"the-line/backend/internal/repository"
)

type ActivityService struct {
	nodeLogRepo *repository.NodeLogRepository
	runRepo     *repository.RunRepository
	runNodeRepo *repository.RunNodeRepository
	personRepo  *repository.PersonRepository
	agentRepo   *repository.AgentRepository
}

func NewActivityService(
	nodeLogRepo *repository.NodeLogRepository,
	runRepo *repository.RunRepository,
	runNodeRepo *repository.RunNodeRepository,
	personRepo *repository.PersonRepository,
	agentRepo *repository.AgentRepository,
) *ActivityService {
	return &ActivityService{
		nodeLogRepo: nodeLogRepo,
		runRepo:     runRepo,
		runNodeRepo: runNodeRepo,
		personRepo:  personRepo,
		agentRepo:   agentRepo,
	}
}

func (s *ActivityService) Recent(ctx context.Context, req dto.RecentActivityRequest, actor domain.Actor) ([]dto.RecentActivityResponse, error) {
	logs, err := s.nodeLogRepo.ListRecent(ctx, req.Limit)
	if err != nil {
		return nil, err
	}

	runMap, nodeMap, personMap, agentMap, err := s.loadActivityRelationMaps(ctx, logs)
	if err != nil {
		return nil, err
	}

	items := make([]dto.RecentActivityResponse, 0, len(logs))
	for _, log := range logs {
		item := dto.RecentActivityResponse{
			ID:           log.ID,
			RunID:        log.RunID,
			RunNodeID:    log.RunNodeID,
			LogType:      log.LogType,
			OperatorType: log.OperatorType,
			OperatorID:   log.OperatorID,
			Content:      log.Content,
			CreatedAt:    log.CreatedAt,
		}
		if run, ok := runMap[log.RunID]; ok {
			item.RunTitle = run.Title
		}
		if node, ok := nodeMap[log.RunNodeID]; ok {
			item.NodeName = node.NodeName
		}
		switch log.OperatorType {
		case domain.OperatorTypePerson:
			if person, ok := personMap[log.OperatorID]; ok {
				item.OperatorName = person.Name
			}
		case domain.OperatorTypeAgent:
			if agent, ok := agentMap[log.OperatorID]; ok {
				item.OperatorName = agent.Name
			}
		case domain.OperatorTypeSystem:
			item.OperatorName = "系统"
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *ActivityService) loadActivityRelationMaps(ctx context.Context, logs []model.FlowRunNodeLog) (map[uint64]model.FlowRun, map[uint64]model.FlowRunNode, map[uint64]model.Person, map[uint64]model.Agent, error) {
	runIDs := make([]uint64, 0)
	nodeIDs := make([]uint64, 0)
	personIDs := make([]uint64, 0)
	agentIDs := make([]uint64, 0)
	seenRunIDs := map[uint64]struct{}{}
	seenNodeIDs := map[uint64]struct{}{}
	seenPersonIDs := map[uint64]struct{}{}
	seenAgentIDs := map[uint64]struct{}{}

	for _, log := range logs {
		if log.RunID > 0 {
			if _, ok := seenRunIDs[log.RunID]; !ok {
				seenRunIDs[log.RunID] = struct{}{}
				runIDs = append(runIDs, log.RunID)
			}
		}
		if log.RunNodeID > 0 {
			if _, ok := seenNodeIDs[log.RunNodeID]; !ok {
				seenNodeIDs[log.RunNodeID] = struct{}{}
				nodeIDs = append(nodeIDs, log.RunNodeID)
			}
		}
		if log.OperatorID > 0 && log.OperatorType == domain.OperatorTypePerson {
			if _, ok := seenPersonIDs[log.OperatorID]; !ok {
				seenPersonIDs[log.OperatorID] = struct{}{}
				personIDs = append(personIDs, log.OperatorID)
			}
		}
		if log.OperatorID > 0 && log.OperatorType == domain.OperatorTypeAgent {
			if _, ok := seenAgentIDs[log.OperatorID]; !ok {
				seenAgentIDs[log.OperatorID] = struct{}{}
				agentIDs = append(agentIDs, log.OperatorID)
			}
		}
	}

	runs, err := s.runRepo.GetByIDs(ctx, runIDs)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	nodes, err := s.runNodeRepo.GetByIDs(ctx, nodeIDs)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	persons, err := s.personRepo.GetByIDs(ctx, personIDs)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	agents, err := s.agentRepo.GetByIDs(ctx, agentIDs)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	runMap := make(map[uint64]model.FlowRun, len(runs))
	for _, run := range runs {
		runMap[run.ID] = run
	}
	nodeMap := make(map[uint64]model.FlowRunNode, len(nodes))
	for _, node := range nodes {
		nodeMap[node.ID] = node
	}
	personMap := make(map[uint64]model.Person, len(persons))
	for _, person := range persons {
		personMap[person.ID] = person
	}
	agentMap := make(map[uint64]model.Agent, len(agents))
	for _, agent := range agents {
		agentMap[agent.ID] = agent
	}
	return runMap, nodeMap, personMap, agentMap, nil
}
