package logic

import (
	"context"

	"cscan/api/internal/svc"
	"cscan/api/internal/types"
	"cscan/model"

	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson"
)

// OrganizationListLogic — список организаций
type OrganizationListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrganizationListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrganizationListLogic {
	return &OrganizationListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OrganizationListLogic) OrganizationList(req *types.PageReq) (resp *types.OrganizationListResp, err error) {
	filter := bson.M{}

	total, err := l.svcCtx.OrganizationModel.Count(l.ctx, filter)
	if err != nil {
		return &types.OrganizationListResp{Code: 500, Msg: "Ошибка получения списка"}, nil
	}

	orgs, err := l.svcCtx.OrganizationModel.Find(l.ctx, filter, req.Page, req.PageSize)
	if err != nil {
		return &types.OrganizationListResp{Code: 500, Msg: "Ошибка получения списка"}, nil
	}

	list := make([]types.Organization, 0, len(orgs))
	for _, o := range orgs {
		list = append(list, types.Organization{
			Id:          o.Id.Hex(),
			Name:        o.Name,
			Description: o.Description,
			Status:      o.Status,
			CreateTime:  o.CreateTime.Local().Format("2006-01-02 15:04:05"),
		})
	}

	return &types.OrganizationListResp{
		Code:  0,
		Msg:   "success",
		Total: int(total),
		List:  list,
	}, nil
}

// OrganizationSaveLogic — создание/обновление организации
type OrganizationSaveLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrganizationSaveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrganizationSaveLogic {
	return &OrganizationSaveLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OrganizationSaveLogic) OrganizationSave(req *types.OrganizationSaveReq) (resp *types.BaseResp, err error) {
	if req.Name == "" {
		return &types.BaseResp{Code: 400, Msg: "Название организации обязательно"}, nil
	}

	if req.Id != "" {
		update := bson.M{
			"name":        req.Name,
			"description": req.Description,
		}
		if req.Status != "" {
			update["status"] = req.Status
		}
		err = l.svcCtx.OrganizationModel.Update(l.ctx, req.Id, update)
		if err != nil {
			return &types.BaseResp{Code: 500, Msg: "Ошибка обновления"}, nil
		}
		return &types.BaseResp{Code: 0, Msg: "Организация обновлена"}, nil
	}

	org := &model.Organization{
		Name:        req.Name,
		Description: req.Description,
	}
	if err = l.svcCtx.OrganizationModel.Insert(l.ctx, org); err != nil {
		return &types.BaseResp{Code: 500, Msg: "Ошибка создания"}, nil
	}

	return &types.BaseResp{Code: 0, Msg: "Организация создана"}, nil
}

// OrganizationDeleteLogic — удаление организации (каскадно — все таргеты)
type OrganizationDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrganizationDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrganizationDeleteLogic {
	return &OrganizationDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OrganizationDeleteLogic) OrganizationDelete(req *types.OrganizationDeleteReq) (resp *types.BaseResp, err error) {
	if req.Id == "" {
		return &types.BaseResp{Code: 400, Msg: "ID не может быть пустым"}, nil
	}

	if err = l.svcCtx.OrganizationModel.Delete(l.ctx, req.Id); err != nil {
		return &types.BaseResp{Code: 500, Msg: "Ошибка удаления"}, nil
	}

	// Каскадно удаляем все таргеты этой организации
	if l.svcCtx.OrgTargetModel != nil {
		if err := l.svcCtx.OrgTargetModel.DeleteByOrg(l.ctx, req.Id); err != nil {
			l.Errorf("Ошибка удаления таргетов организации %s: %v", req.Id, err)
		}
	}

	return &types.BaseResp{Code: 0, Msg: "Организация удалена"}, nil
}

// OrganizationUpdateStatusLogic — изменение статуса организации
type OrganizationUpdateStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrganizationUpdateStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrganizationUpdateStatusLogic {
	return &OrganizationUpdateStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OrganizationUpdateStatusLogic) OrganizationUpdateStatus(req *types.OrganizationUpdateStatusReq) (resp *types.BaseResp, err error) {
	if req.Id == "" {
		return &types.BaseResp{Code: 400, Msg: "ID не может быть пустым"}, nil
	}
	if req.Status == "" {
		return &types.BaseResp{Code: 400, Msg: "Статус не может быть пустым"}, nil
	}

	err = l.svcCtx.OrganizationModel.Update(l.ctx, req.Id, bson.M{"status": req.Status})
	if err != nil {
		return &types.BaseResp{Code: 500, Msg: "Ошибка обновления статуса"}, nil
	}

	return &types.BaseResp{Code: 0, Msg: "Статус обновлён"}, nil
}
