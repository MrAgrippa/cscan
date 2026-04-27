package logic

import (
	"context"
	"strings"

	"cscan/api/internal/svc"
	"cscan/api/internal/types"
	"cscan/model"

	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson"
)

// =====================================================================
// OrgTargetListLogic — список таргетов организации
// =====================================================================
type OrgTargetListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrgTargetListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrgTargetListLogic {
	return &OrgTargetListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OrgTargetListLogic) OrgTargetList(req *types.OrgTargetListReq) (*types.OrgTargetListResp, error) {
	if req.OrgId == "" {
		return &types.OrgTargetListResp{Code: 400, Msg: "orgId обязателен"}, nil
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	docs, total, err := l.svcCtx.OrgTargetModel.FindPaged(l.ctx, req.OrgId, req.Search, req.Type, req.Page, req.PageSize)
	if err != nil {
		l.Errorf("OrgTargetList: %v", err)
		return &types.OrgTargetListResp{Code: 500, Msg: "Ошибка получения списка"}, nil
	}

	list := make([]types.OrgTarget, 0, len(docs))
	for _, d := range docs {
		list = append(list, types.OrgTarget{
			Id:          d.Id.Hex(),
			OrgId:       d.OrgId,
			Type:        d.Type,
			Value:       d.Value,
			Description: d.Description,
			Enabled:     d.Enabled,
			CreateTime:  d.CreateTime.Local().Format("2006-01-02 15:04:05"),
		})
	}

	return &types.OrgTargetListResp{
		Code:  0,
		Msg:   "success",
		Total: int(total),
		List:  list,
	}, nil
}

// =====================================================================
// OrgTargetSaveLogic — добавление группы таргетов в организацию
// =====================================================================
type OrgTargetSaveLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrgTargetSaveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrgTargetSaveLogic {
	return &OrgTargetSaveLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OrgTargetSaveLogic) OrgTargetSave(req *types.OrgTargetSaveReq) (*types.OrgTargetSaveResp, error) {
	if req.OrgId == "" {
		return &types.OrgTargetSaveResp{Code: 400, Msg: "orgId обязателен"}, nil
	}
	// Проверяем существование организации
	if _, err := l.svcCtx.OrganizationModel.FindById(l.ctx, req.OrgId); err != nil {
		return &types.OrgTargetSaveResp{Code: 404, Msg: "Организация не найдена"}, nil
	}

	items := make([]model.OrgTarget, 0, len(req.Items))
	for _, it := range req.Items {
		v := strings.TrimSpace(it.Value)
		if v == "" {
			continue
		}
		t := it.Type
		if t == "" {
			t = model.DetectOrgTargetType(v)
		}
		items = append(items, model.OrgTarget{
			Type:        t,
			Value:       v,
			Description: it.Description,
		})
	}
	if len(items) == 0 {
		return &types.OrgTargetSaveResp{Code: 400, Msg: "Список таргетов пуст"}, nil
	}

	inserted, err := l.svcCtx.OrgTargetModel.BulkInsert(l.ctx, req.OrgId, items)
	if err != nil {
		l.Errorf("OrgTargetSave BulkInsert: %v", err)
	}
	skipped := len(items) - inserted

	// Каскадно: пробуем привязать существующие ассеты к новым таргетам
	go l.retagAssetsForOrg(req.OrgId)

	return &types.OrgTargetSaveResp{
		Code:     0,
		Msg:      "Сохранено",
		Inserted: inserted,
		Skipped:  skipped,
	}, nil
}

// retagAssetsForOrg — фоновое назначение org_id ассетам по таргетам организации
func (l *OrgTargetSaveLogic) retagAssetsForOrg(orgId string) {
	bgCtx := context.Background()
	targets, err := l.svcCtx.OrgTargetModel.FindByOrg(bgCtx, orgId)
	if err != nil {
		logx.Errorf("[retagAssetsForOrg] FindByOrg %s: %v", orgId, err)
		return
	}
	if len(targets) == 0 {
		return
	}
	matcher := model.NewOrgTargetMatcher(targets)
	// Список workspace'ов: основной default + все из коллекции workspace
	workspaces := map[string]struct{}{"default": {}}
	if l.svcCtx.WorkspaceModel != nil {
		if list, err := l.svcCtx.WorkspaceModel.Find(bgCtx, bson.M{}, 0, 0); err == nil {
			for _, w := range list {
				if !w.Id.IsZero() {
					workspaces[w.Id.Hex()] = struct{}{}
				}
			}
		}
	}
	for ws := range workspaces {
		am := l.svcCtx.GetAssetModel(ws)
		assets, err := am.FindHostsAndIPs(bgCtx)
		if err != nil {
			logx.Errorf("[retagAssetsForOrg] FindHostsAndIPs %s: %v", ws, err)
			continue
		}
		// Сгруппируем host'ы по orgId, к которому их нужно привязать
		hostsByOrg := make(map[string][]string)
		for _, a := range assets {
			matchedOrg := matcher.Match(a.Host)
			if matchedOrg == "" && a.Domain != "" {
				matchedOrg = matcher.Match(a.Domain)
			}
			if matchedOrg == orgId {
				// привязываем только к ЭТОЙ организации
				hostsByOrg[orgId] = append(hostsByOrg[orgId], a.Host)
			}
		}
		hosts := hostsByOrg[orgId]
		if len(hosts) == 0 {
			continue
		}
		_, _ = am.SetOrgIdByHosts(bgCtx, hosts, orgId)
		// Каскадно — vul и dirscan
		_, _ = l.svcCtx.GetVulModel(ws).SetOrgIdByHosts(bgCtx, hosts, orgId)
		_, _ = l.svcCtx.GetDirScanResultModel().SetOrgIdByHosts(bgCtx, hosts, orgId)
	}
}

// =====================================================================
// OrgTargetImportLogic — массовый импорт таргетов из текста
// =====================================================================
type OrgTargetImportLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrgTargetImportLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrgTargetImportLogic {
	return &OrgTargetImportLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OrgTargetImportLogic) OrgTargetImport(req *types.OrgTargetImportReq) (*types.OrgTargetSaveResp, error) {
	if req.OrgId == "" {
		return &types.OrgTargetSaveResp{Code: 400, Msg: "orgId обязателен"}, nil
	}
	// Парсим текст: разделители — \n, запятая, точка с запятой, пробел
	rawTokens := strings.FieldsFunc(req.Text, func(r rune) bool {
		return r == '\n' || r == ',' || r == ';' || r == ' ' || r == '\t' || r == '\r'
	})
	items := make([]model.OrgTarget, 0, len(rawTokens))
	for _, tok := range rawTokens {
		v := strings.TrimSpace(tok)
		if v == "" {
			continue
		}
		items = append(items, model.OrgTarget{
			Type:  model.DetectOrgTargetType(v),
			Value: v,
		})
	}
	if len(items) == 0 {
		return &types.OrgTargetSaveResp{Code: 400, Msg: "Не найдено валидных таргетов"}, nil
	}

	inserted, err := l.svcCtx.OrgTargetModel.BulkInsert(l.ctx, req.OrgId, items)
	if err != nil {
		l.Errorf("OrgTargetImport BulkInsert: %v", err)
	}

	// Запускаем retag в фоне
	go (&OrgTargetSaveLogic{ctx: context.Background(), svcCtx: l.svcCtx, Logger: l.Logger}).retagAssetsForOrg(req.OrgId)

	return &types.OrgTargetSaveResp{
		Code:     0,
		Msg:      "Импортировано",
		Inserted: inserted,
		Skipped:  len(items) - inserted,
	}, nil
}

// =====================================================================
// OrgTargetUpdateLogic — обновление одного таргета
// =====================================================================
type OrgTargetUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrgTargetUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrgTargetUpdateLogic {
	return &OrgTargetUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OrgTargetUpdateLogic) OrgTargetUpdate(req *types.OrgTargetUpdateReq) (*types.BaseResp, error) {
	if req.Id == "" {
		return &types.BaseResp{Code: 400, Msg: "id обязателен"}, nil
	}
	update := bson.M{}
	if req.Value != "" {
		t := req.Type
		if t == "" {
			t = model.DetectOrgTargetType(req.Value)
		}
		update["value"] = model.NormalizeOrgTargetValue(req.Value, t)
		update["type"] = t
	} else if req.Type != "" {
		update["type"] = req.Type
	}
	if req.Description != "" {
		update["description"] = req.Description
	}
	if req.Enabled != nil {
		update["enabled"] = *req.Enabled
	}
	if len(update) == 0 {
		return &types.BaseResp{Code: 400, Msg: "Нет полей для обновления"}, nil
	}
	if err := l.svcCtx.OrgTargetModel.Update(l.ctx, req.Id, update); err != nil {
		l.Errorf("OrgTargetUpdate: %v", err)
		return &types.BaseResp{Code: 500, Msg: "Ошибка обновления"}, nil
	}
	return &types.BaseResp{Code: 0, Msg: "Обновлено"}, nil
}

// =====================================================================
// OrgTargetDeleteLogic — удаление одного или нескольких таргетов
// =====================================================================
type OrgTargetDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrgTargetDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrgTargetDeleteLogic {
	return &OrgTargetDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OrgTargetDeleteLogic) OrgTargetDelete(req *types.OrgTargetDeleteReq) (*types.BaseResp, error) {
	if len(req.Ids) == 0 {
		return &types.BaseResp{Code: 400, Msg: "Список id пуст"}, nil
	}
	deleted, err := l.svcCtx.OrgTargetModel.DeleteMany(l.ctx, req.Ids)
	if err != nil {
		l.Errorf("OrgTargetDelete: %v", err)
		return &types.BaseResp{Code: 500, Msg: "Ошибка удаления"}, nil
	}
	return &types.BaseResp{Code: 0, Msg: "Удалено: " + intToStr(int64(deleted))}, nil
}

// =====================================================================
// OrgAssignAssetsLogic — ручная привязка ассетов к организации
// =====================================================================
type OrgAssignAssetsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrgAssignAssetsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrgAssignAssetsLogic {
	return &OrgAssignAssetsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OrgAssignAssetsLogic) OrgAssignAssets(req *types.OrgAssignAssetsReq) (*types.OrgAssignAssetsResp, error) {
	if len(req.AssetIds) == 0 {
		return &types.OrgAssignAssetsResp{Code: 400, Msg: "Список ассетов пуст"}, nil
	}
	ws := req.WorkspaceId
	if ws == "" {
		ws = "default"
	}
	am := l.svcCtx.GetAssetModel(ws)
	modified, err := am.SetOrgIdByIds(l.ctx, req.AssetIds, req.OrgId)
	if err != nil {
		l.Errorf("OrgAssignAssets: %v", err)
		return &types.OrgAssignAssetsResp{Code: 500, Msg: "Ошибка привязки"}, nil
	}

	// Каскадно — vul и dirscan
	if req.OrgId != "" {
		// собрать host'ы привязываемых ассетов
		hosts := make([]string, 0, len(req.AssetIds))
		for _, id := range req.AssetIds {
			a, err := am.FindById(l.ctx, id)
			if err == nil && a != nil {
				hosts = append(hosts, a.Host)
			}
		}
		if len(hosts) > 0 {
			_, _ = l.svcCtx.GetVulModel(ws).SetOrgIdByHosts(l.ctx, hosts, req.OrgId)
			_, _ = l.svcCtx.GetDirScanResultModel().SetOrgIdByHosts(l.ctx, hosts, req.OrgId)
		}
	}

	msg := "Привязано"
	if req.OrgId == "" {
		msg = "Отвязано"
	}
	return &types.OrgAssignAssetsResp{Code: 0, Msg: msg, Modified: int(modified)}, nil
}

// =====================================================================
// OrgRetagLogic — пересборка привязки всех ассетов по таргетам организаций
// =====================================================================
type OrgRetagLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrgRetagLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrgRetagLogic {
	return &OrgRetagLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OrgRetagLogic) OrgRetag(req *types.OrgRetagReq) (*types.OrgRetagResp, error) {
	ws := req.WorkspaceId
	if ws == "" {
		ws = "default"
	}

	// Загружаем все таргеты (или для одной организации)
	var targets []model.OrgTarget
	var err error
	if req.OrgId != "" {
		targets, err = l.svcCtx.OrgTargetModel.FindByOrg(l.ctx, req.OrgId)
	} else {
		targets, err = l.svcCtx.OrgTargetModel.AllEnabled(l.ctx)
	}
	if err != nil {
		l.Errorf("OrgRetag: %v", err)
		return &types.OrgRetagResp{Code: 500, Msg: "Ошибка чтения таргетов"}, nil
	}
	matcher := model.NewOrgTargetMatcher(targets)

	am := l.svcCtx.GetAssetModel(ws)
	assets, err := am.FindHostsAndIPs(l.ctx)
	if err != nil {
		l.Errorf("OrgRetag find assets: %v", err)
		return &types.OrgRetagResp{Code: 500, Msg: "Ошибка чтения ассетов"}, nil
	}

	// orgId -> список host'ов
	hostsByOrg := make(map[string][]string)
	matched := 0
	for _, a := range assets {
		match := matcher.Match(a.Host)
		if match == "" && a.Domain != "" {
			match = matcher.Match(a.Domain)
		}
		if match != "" {
			hostsByOrg[match] = append(hostsByOrg[match], a.Host)
			matched++
		}
	}

	updated := 0
	for orgId, hosts := range hostsByOrg {
		if len(hosts) == 0 {
			continue
		}
		mod, err := am.SetOrgIdByHosts(l.ctx, hosts, orgId)
		if err != nil {
			l.Errorf("OrgRetag SetOrgIdByHosts: %v", err)
			continue
		}
		updated += int(mod)
		// Каскад
		_, _ = l.svcCtx.GetVulModel(ws).SetOrgIdByHosts(l.ctx, hosts, orgId)
		_, _ = l.svcCtx.GetDirScanResultModel().SetOrgIdByHosts(l.ctx, hosts, orgId)
	}

	return &types.OrgRetagResp{
		Code:    0,
		Msg:     "Готово",
		Matched: matched,
		Updated: updated,
	}, nil
}

func intToStr(n int64) string {
	const digits = "0123456789"
	if n == 0 {
		return "0"
	}
	negative := n < 0
	if negative {
		n = -n
	}
	buf := make([]byte, 0, 20)
	for n > 0 {
		buf = append([]byte{digits[n%10]}, buf...)
		n /= 10
	}
	if negative {
		return "-" + string(buf)
	}
	return string(buf)
}
