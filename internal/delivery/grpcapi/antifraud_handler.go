package grpcapi

import (
    "context"

    "github.com/LavaJover/shvark-order-service/internal/domain"
    "github.com/LavaJover/shvark-order-service/internal/usecase"
    antifraudpb "github.com/LavaJover/shvark-order-service/proto/gen"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "google.golang.org/protobuf/types/known/structpb"
    "google.golang.org/protobuf/types/known/timestamppb"
)

type AntiFraudHandler struct {
    antifraudpb.UnimplementedAntiFraudServiceServer
    useCase usecase.AntiFraudUseCase
}

func NewAntiFraudHandler(useCase usecase.AntiFraudUseCase) *AntiFraudHandler {
    return &AntiFraudHandler{
        useCase: useCase,
    }
}

// ============= Проверка трейдера =============

func (h *AntiFraudHandler) CheckTrader(ctx context.Context, req *antifraudpb.CheckTraderRequest) (*antifraudpb.CheckTraderResponse, error) {
    if req.TraderId == "" {
        return nil, status.Error(codes.InvalidArgument, "trader_id is required")
    }

    report, err := h.useCase.CheckTrader(ctx, req.TraderId)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to check trader: %v", err)
    }

    results := make([]*antifraudpb.CheckResult, 0, len(report.Results))
    for _, r := range report.Results {
        details, _ := structpb.NewStruct(r.Details)
        results = append(results, &antifraudpb.CheckResult{
            RuleName: r.RuleName,
            Passed:   r.Passed,
            Message:  r.Message,
            Details:  details,
        })
    }

    return &antifraudpb.CheckTraderResponse{
        TraderId:    report.TraderID,
        CheckedAt:   timestamppb.New(report.CheckedAt),
        AllPassed:   report.AllPassed,
        Results:     results,
        FailedRules: report.FailedRules,
    }, nil
}

func (h *AntiFraudHandler) ProcessTraderCheck(ctx context.Context, req *antifraudpb.ProcessTraderCheckRequest) (*antifraudpb.ProcessTraderCheckResponse, error) {
    if req.TraderId == "" {
        return nil, status.Error(codes.InvalidArgument, "trader_id is required")
    }

    err := h.useCase.ProcessTraderCheck(ctx, req.TraderId)
    if err != nil {
        return &antifraudpb.ProcessTraderCheckResponse{
            Success: false,
            Message: err.Error(),
        }, nil
    }

    return &antifraudpb.ProcessTraderCheckResponse{
        Success: true,
        Message: "Trader check processed successfully",
    }, nil
}

// ============= Управление правилами =============

func (h *AntiFraudHandler) CreateRule(ctx context.Context, req *antifraudpb.CreateRuleRequest) (*antifraudpb.CreateRuleResponse, error) {
    if req.Name == "" {
        return nil, status.Error(codes.InvalidArgument, "name is required")
    }
    if req.Type == "" {
        return nil, status.Error(codes.InvalidArgument, "type is required")
    }
    if req.Config == nil {
        return nil, status.Error(codes.InvalidArgument, "config is required")
    }

    config := req.Config.AsMap()

    domainReq := &domain.CreateRuleRequest{
        Name:     req.Name,
        Type:     req.Type,
        Config:   config,
        Priority: int(req.Priority),
    }

    rule, err := h.useCase.CreateRule(ctx, domainReq)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to create rule: %v", err)
    }

    return &antifraudpb.CreateRuleResponse{
        Rule: h.convertDomainRuleToProto(rule),
    }, nil
}

func (h *AntiFraudHandler) UpdateRule(ctx context.Context, req *antifraudpb.UpdateRuleRequest) (*antifraudpb.UpdateRuleResponse, error) {
    if req.RuleId == "" {
        return nil, status.Error(codes.InvalidArgument, "rule_id is required")
    }

    domainReq := &domain.UpdateRuleRequest{
        RuleID: req.RuleId,
    }

    if req.Config != nil {
        domainReq.Config = req.Config.AsMap()
    }

    if req.IsActive != nil {
        isActive := *req.IsActive
        domainReq.IsActive = &isActive
    }

    if req.Priority != nil {
        priority := int(*req.Priority)
        domainReq.Priority = &priority
    }

    err := h.useCase.UpdateRule(ctx, domainReq)
    if err != nil {
        return &antifraudpb.UpdateRuleResponse{
            Success: false,
            Message: err.Error(),
        }, nil
    }

    return &antifraudpb.UpdateRuleResponse{
        Success: true,
        Message: "Rule updated successfully",
    }, nil
}

func (h *AntiFraudHandler) GetRules(ctx context.Context, req *antifraudpb.GetRulesRequest) (*antifraudpb.GetRulesResponse, error) {
    rules, err := h.useCase.GetRules(ctx, req.ActiveOnly)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to get rules: %v", err)
    }

    protoRules := make([]*antifraudpb.AntiFraudRule, 0, len(rules))
    for _, rule := range rules {
        protoRules = append(protoRules, h.convertDomainRuleToProto(rule))
    }

    return &antifraudpb.GetRulesResponse{
        Rules: protoRules,
    }, nil
}

func (h *AntiFraudHandler) GetRule(ctx context.Context, req *antifraudpb.GetRuleRequest) (*antifraudpb.GetRuleResponse, error) {
    if req.RuleId == "" {
        return nil, status.Error(codes.InvalidArgument, "rule_id is required")
    }

    rule, err := h.useCase.GetRule(ctx, req.RuleId)
    if err != nil {
        return nil, status.Errorf(codes.NotFound, "rule not found: %v", err)
    }

    return &antifraudpb.GetRuleResponse{
        Rule: h.convertDomainRuleToProto(rule),
    }, nil
}

func (h *AntiFraudHandler) DeleteRule(ctx context.Context, req *antifraudpb.DeleteRuleRequest) (*antifraudpb.DeleteRuleResponse, error) {
    if req.RuleId == "" {
        return nil, status.Error(codes.InvalidArgument, "rule_id is required")
    }

    err := h.useCase.DeleteRule(ctx, req.RuleId)
    if err != nil {
        return &antifraudpb.DeleteRuleResponse{
            Success: false,
            Message: err.Error(),
        }, nil
    }

    return &antifraudpb.DeleteRuleResponse{
        Success: true,
        Message: "Rule deleted successfully",
    }, nil
}

// ============= Аудит =============

func (h *AntiFraudHandler) GetAuditLogs(ctx context.Context, req *antifraudpb.GetAuditLogsRequest) (*antifraudpb.GetAuditLogsResponse, error) {
    domainReq := &domain.GetAuditLogsRequest{
        OnlyFailed: req.OnlyFailed,
        Limit:      int(req.Limit),
        Offset:     int(req.Offset),
    }

    if req.TraderId != nil {
        domainReq.TraderID = *req.TraderId
    }

    if req.FromDate != nil {
        fromDate := req.FromDate.AsTime()
        domainReq.FromDate = &fromDate
    }

    if req.ToDate != nil {
        toDate := req.ToDate.AsTime()
        domainReq.ToDate = &toDate
    }

    logs, err := h.useCase.GetAuditLogs(ctx, domainReq)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to get audit logs: %v", err)
    }

    protoLogs := make([]*antifraudpb.AuditLog, 0, len(logs))
    for _, log := range logs {
        protoLogs = append(protoLogs, h.convertDomainAuditLogToProto(log))
    }

    return &antifraudpb.GetAuditLogsResponse{
        Logs:  protoLogs,
        Total: int32(len(protoLogs)),
    }, nil
}

func (h *AntiFraudHandler) GetTraderAuditHistory(ctx context.Context, req *antifraudpb.GetTraderAuditHistoryRequest) (*antifraudpb.GetTraderAuditHistoryResponse, error) {
    if req.TraderId == "" {
        return nil, status.Error(codes.InvalidArgument, "trader_id is required")
    }

    limit := int(req.Limit)
    if limit <= 0 {
        limit = 10
    }

    logs, err := h.useCase.GetTraderAuditHistory(ctx, req.TraderId, limit)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to get trader audit history: %v", err)
    }

    protoLogs := make([]*antifraudpb.AuditLog, 0, len(logs))
    for _, log := range logs {
        protoLogs = append(protoLogs, h.convertDomainAuditLogToProto(log))
    }

    return &antifraudpb.GetTraderAuditHistoryResponse{
        Logs: protoLogs,
    }, nil
}

// ============= Вспомогательные функции =============

func (h *AntiFraudHandler) convertDomainRuleToProto(rule *domain.AntiFraudRuleResponse) *antifraudpb.AntiFraudRule {
    config, _ := structpb.NewStruct(rule.Config)

    return &antifraudpb.AntiFraudRule{
        Id:        rule.ID,
        Name:      rule.Name,
        Type:      rule.Type,
        Config:    config,
        IsActive:  rule.IsActive,
        Priority:  int32(rule.Priority),
        CreatedAt: timestamppb.New(rule.CreatedAt),
        UpdatedAt: timestamppb.New(rule.UpdatedAt),
    }
}

func (h *AntiFraudHandler) convertDomainAuditLogToProto(log *domain.AuditLogResponse) *antifraudpb.AuditLog {
    results := make([]*antifraudpb.CheckResult, 0, len(log.Results))
    for _, r := range log.Results {
        details, _ := structpb.NewStruct(r.Details)
        results = append(results, &antifraudpb.CheckResult{
            RuleName: r.RuleName,
            Passed:   r.Passed,
            Message:  r.Message,
            Details:  details,
        })
    }

    return &antifraudpb.AuditLog{
        Id:        log.ID,
        TraderId:  log.TraderID,
        CheckedAt: timestamppb.New(log.CheckedAt),
        AllPassed: log.AllPassed,
        Results:   results,
        CreatedAt: timestamppb.New(log.CreatedAt),
    }
}

// ManualUnlock вручную разблокирует трейдера
func (h *AntiFraudHandler) ManualUnlock(ctx context.Context, req *antifraudpb.ManualUnlockRequest) (*antifraudpb.ManualUnlockResponse, error) {
    if req.TraderId == "" {
        return nil, status.Error(codes.InvalidArgument, "trader_id is required")
    }

    if req.AdminId == "" {
        return nil, status.Error(codes.InvalidArgument, "admin_id is required")
    }

    domainReq := &domain.ManualUnlockRequest{
        TraderID:         req.TraderId,
        AdminID:          req.AdminId,
        Reason:           req.Reason,
        GracePeriodHours: int(req.GracePeriodHours),
    }

    if domainReq.GracePeriodHours <= 0 {
        domainReq.GracePeriodHours = 24
    }

    err := h.useCase.ManualUnlock(ctx, domainReq)
    if err != nil {
        return &antifraudpb.ManualUnlockResponse{
            Success: false,
            Message: err.Error(),
        }, nil
    }

    gracePeriodUntil := timestamppb.Now()
    gracePeriodUntil.Seconds += int64(domainReq.GracePeriodHours * 3600)

    return &antifraudpb.ManualUnlockResponse{
        Success:          true,
        Message:          "Trader unlocked successfully",
        GracePeriodUntil: gracePeriodUntil,
    }, nil
}

// ResetGracePeriod сбрасывает грейс-период
func (h *AntiFraudHandler) ResetGracePeriod(ctx context.Context, req *antifraudpb.ResetGracePeriodRequest) (*antifraudpb.ResetGracePeriodResponse, error) {
    if req.TraderId == "" {
        return nil, status.Error(codes.InvalidArgument, "trader_id is required")
    }

    err := h.useCase.ResetGracePeriod(ctx, req.TraderId)
    if err != nil {
        return &antifraudpb.ResetGracePeriodResponse{
            Success: false,
            Message: err.Error(),
        }, nil
    }

    return &antifraudpb.ResetGracePeriodResponse{
        Success: true,
        Message: "Grace period reset successfully",
    }, nil
}

// GetUnlockHistory получает историю разблокировок трейдера
func (h *AntiFraudHandler) GetUnlockHistory(ctx context.Context, req *antifraudpb.GetUnlockHistoryRequest) (*antifraudpb.GetUnlockHistoryResponse, error) {
    if req.TraderId == "" {
        return nil, status.Error(codes.InvalidArgument, "trader_id is required")
    }

    limit := int(req.Limit)
    if limit <= 0 {
        limit = 20
    }

    logs, err := h.useCase.GetUnlockHistory(ctx, req.TraderId, limit)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to get unlock history: %v", err)
    }

    items := make([]*antifraudpb.UnlockHistoryItem, 0, len(logs))
    for _, log := range logs {
        items = append(items, &antifraudpb.UnlockHistoryItem{
            Id:               log.ID,
            TraderId:         log.TraderID,
            AdminId:          log.AdminID,
            Reason:           log.Reason,
            GracePeriodHours: int32(log.GracePeriodHours),
            UnlockedAt:       timestamppb.New(log.UnlockedAt),
            CreatedAt:        timestamppb.New(log.CreatedAt),
        })
    }

    return &antifraudpb.GetUnlockHistoryResponse{
        Items: items,
    }, nil
}