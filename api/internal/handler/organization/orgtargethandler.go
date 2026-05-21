package organization

import (
	"net/http"

	"cscan/api/internal/logic"
	"cscan/api/internal/svc"
	"cscan/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// OrgTargetListHandler — список таргетов организации
func OrgTargetListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.OrgTargetListReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.OkJson(w, &types.BaseResp{Code: 400, Msg: err.Error()})
			return
		}
		l := logic.NewOrgTargetListLogic(r.Context(), svcCtx)
		resp, _ := l.OrgTargetList(&req)
		httpx.OkJson(w, resp)
	}
}

// OrgTargetSaveHandler — добавление группы таргетов
func OrgTargetSaveHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.OrgTargetSaveReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.OkJson(w, &types.BaseResp{Code: 400, Msg: err.Error()})
			return
		}
		l := logic.NewOrgTargetSaveLogic(r.Context(), svcCtx)
		resp, _ := l.OrgTargetSave(&req)
		httpx.OkJson(w, resp)
	}
}

// OrgTargetImportHandler — импорт таргетов из текста
func OrgTargetImportHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.OrgTargetImportReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.OkJson(w, &types.BaseResp{Code: 400, Msg: err.Error()})
			return
		}
		l := logic.NewOrgTargetImportLogic(r.Context(), svcCtx)
		resp, _ := l.OrgTargetImport(&req)
		httpx.OkJson(w, resp)
	}
}

// OrgTargetUpdateHandler — обновление таргета
func OrgTargetUpdateHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.OrgTargetUpdateReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.OkJson(w, &types.BaseResp{Code: 400, Msg: err.Error()})
			return
		}
		l := logic.NewOrgTargetUpdateLogic(r.Context(), svcCtx)
		resp, _ := l.OrgTargetUpdate(&req)
		httpx.OkJson(w, resp)
	}
}

// OrgTargetDeleteHandler — удаление таргета
func OrgTargetDeleteHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.OrgTargetDeleteReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.OkJson(w, &types.BaseResp{Code: 400, Msg: err.Error()})
			return
		}
		l := logic.NewOrgTargetDeleteLogic(r.Context(), svcCtx)
		resp, _ := l.OrgTargetDelete(&req)
		httpx.OkJson(w, resp)
	}
}

// OrgAssignAssetsHandler — ручная привязка/отвязка ассетов к организации
func OrgAssignAssetsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.OrgAssignAssetsReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.OkJson(w, &types.BaseResp{Code: 400, Msg: err.Error()})
			return
		}
		l := logic.NewOrgAssignAssetsLogic(r.Context(), svcCtx)
		resp, _ := l.OrgAssignAssets(&req)
		httpx.OkJson(w, resp)
	}
}

// OrgRetagHandler — пересборка привязки ассетов к организациям по таргетам
func OrgRetagHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.OrgRetagReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.OkJson(w, &types.BaseResp{Code: 400, Msg: err.Error()})
			return
		}
		l := logic.NewOrgRetagLogic(r.Context(), svcCtx)
		resp, _ := l.OrgRetag(&req)
		httpx.OkJson(w, resp)
	}
}
