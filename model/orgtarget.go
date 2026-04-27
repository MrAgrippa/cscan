package model

import (
	"context"
	"net"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// OrgTargetType — тип таргета организации
const (
	OrgTargetTypeIP       = "ip"       // 1.2.3.4
	OrgTargetTypeCIDR     = "cidr"     // 10.0.0.0/24
	OrgTargetTypeDomain   = "domain"   // example.com
	OrgTargetTypeWildcard = "wildcard" // *.example.com
)

// OrgTarget — таргет (цель сканирования), привязанный к организации.
// Используется для:
//  1. Авто-привязки сканируемых ассетов к организации (matcher по host/IP)
//  2. Авто-подстановки списка целей в форму создания задачи
//  3. Реестра ресурсов организации
type OrgTarget struct {
	Id          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OrgId       string             `bson:"org_id" json:"orgId"`
	Type        string             `bson:"type" json:"type"`               // ip | cidr | domain | wildcard
	Value       string             `bson:"value" json:"value"`             // сама строка таргета
	Description string             `bson:"description" json:"description"` // комментарий
	Enabled     bool               `bson:"enabled" json:"enabled"`         // true = используется для матчинга и сканирования
	CreateTime  time.Time          `bson:"create_time" json:"createTime"`
	UpdateTime  time.Time          `bson:"update_time" json:"updateTime"`
}

type OrgTargetModel struct {
	coll *mongo.Collection
}

func NewOrgTargetModel(db *mongo.Database) *OrgTargetModel {
	return &OrgTargetModel{
		coll: db.Collection("org_target"),
	}
}

// EnsureIndexes — создаёт индексы для коллекции org_target
func (m *OrgTargetModel) EnsureIndexes(ctx context.Context) error {
	idxs := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "org_id", Value: 1}},
			Options: options.Index().SetBackground(true).SetName("idx_org_id"),
		},
		{
			Keys:    bson.D{{Key: "org_id", Value: 1}, {Key: "value", Value: 1}},
			Options: options.Index().SetUnique(true).SetBackground(true).SetName("idx_org_value_unique"),
		},
		{
			Keys:    bson.D{{Key: "type", Value: 1}},
			Options: options.Index().SetBackground(true).SetName("idx_type"),
		},
		{
			Keys:    bson.D{{Key: "value", Value: 1}},
			Options: options.Index().SetBackground(true).SetName("idx_value"),
		},
	}
	_, err := m.coll.Indexes().CreateMany(ctx, idxs)
	if err != nil && !strings.Contains(err.Error(), "already exists") &&
		!strings.Contains(err.Error(), "IndexOptionsConflict") {
		return err
	}
	return nil
}

// DetectType — автоматически определяет тип таргета по строке
func DetectOrgTargetType(value string) string {
	v := strings.TrimSpace(value)
	if v == "" {
		return ""
	}
	if strings.HasPrefix(v, "*.") {
		return OrgTargetTypeWildcard
	}
	if strings.Contains(v, "/") {
		if _, _, err := net.ParseCIDR(v); err == nil {
			return OrgTargetTypeCIDR
		}
	}
	if ip := net.ParseIP(v); ip != nil {
		return OrgTargetTypeIP
	}
	return OrgTargetTypeDomain
}

// NormalizeValue — нормализует таргет (lowercase для доменов, обрезка пробелов)
func NormalizeOrgTargetValue(value, t string) string {
	v := strings.TrimSpace(value)
	switch t {
	case OrgTargetTypeDomain, OrgTargetTypeWildcard:
		return strings.ToLower(v)
	default:
		return v
	}
}

// Insert — добавляет один таргет
func (m *OrgTargetModel) Insert(ctx context.Context, doc *OrgTarget) error {
	if doc.Id.IsZero() {
		doc.Id = primitive.NewObjectID()
	}
	if doc.Type == "" {
		doc.Type = DetectOrgTargetType(doc.Value)
	}
	doc.Value = NormalizeOrgTargetValue(doc.Value, doc.Type)
	now := time.Now()
	doc.CreateTime = now
	doc.UpdateTime = now
	_, err := m.coll.InsertOne(ctx, doc)
	return err
}

// BulkInsert — импорт списка таргетов; уже существующие (по org_id+value) пропускаются
func (m *OrgTargetModel) BulkInsert(ctx context.Context, orgId string, items []OrgTarget) (int, error) {
	if orgId == "" || len(items) == 0 {
		return 0, nil
	}
	now := time.Now()
	docs := make([]interface{}, 0, len(items))
	for i := range items {
		t := items[i].Type
		if t == "" {
			t = DetectOrgTargetType(items[i].Value)
		}
		v := NormalizeOrgTargetValue(items[i].Value, t)
		if v == "" {
			continue
		}
		docs = append(docs, OrgTarget{
			Id:          primitive.NewObjectID(),
			OrgId:       orgId,
			Type:        t,
			Value:       v,
			Description: items[i].Description,
			Enabled:     true,
			CreateTime:  now,
			UpdateTime:  now,
		})
	}
	if len(docs) == 0 {
		return 0, nil
	}
	opts := options.InsertMany().SetOrdered(false)
	res, err := m.coll.InsertMany(ctx, docs, opts)
	inserted := 0
	if res != nil {
		inserted = len(res.InsertedIDs)
	}
	// Игнорируем дубликаты — они ожидаемы при повторном импорте
	if err != nil {
		if writeEx, ok := err.(mongo.BulkWriteException); ok {
			allDup := true
			for _, we := range writeEx.WriteErrors {
				if we.Code != 11000 {
					allDup = false
					break
				}
			}
			if allDup {
				return inserted, nil
			}
		}
		return inserted, err
	}
	return inserted, nil
}

// FindByOrg — все таргеты организации
func (m *OrgTargetModel) FindByOrg(ctx context.Context, orgId string) ([]OrgTarget, error) {
	cursor, err := m.coll.Find(ctx, bson.M{"org_id": orgId},
		options.Find().SetSort(bson.D{{Key: "create_time", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var docs []OrgTarget
	if err = cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}

// FindPaged — постраничный список таргетов организации с поиском
func (m *OrgTargetModel) FindPaged(ctx context.Context, orgId, search, targetType string, page, pageSize int) ([]OrgTarget, int64, error) {
	filter := bson.M{"org_id": orgId}
	if targetType != "" {
		filter["type"] = targetType
	}
	if search != "" {
		filter["value"] = bson.M{"$regex": primitive.Regex{Pattern: search, Options: "i"}}
	}
	total, err := m.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	opts := options.Find().SetSort(bson.D{{Key: "create_time", Value: -1}})
	if page > 0 && pageSize > 0 {
		opts.SetSkip(int64((page - 1) * pageSize))
		opts.SetLimit(int64(pageSize))
	}
	cursor, err := m.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)
	var docs []OrgTarget
	if err = cursor.All(ctx, &docs); err != nil {
		return nil, 0, err
	}
	return docs, total, nil
}

// Delete — удаляет один таргет по id
func (m *OrgTargetModel) Delete(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = m.coll.DeleteOne(ctx, bson.M{"_id": oid})
	return err
}

// DeleteByOrg — удаляет все таргеты организации (вызывается при удалении самой организации)
func (m *OrgTargetModel) DeleteByOrg(ctx context.Context, orgId string) error {
	_, err := m.coll.DeleteMany(ctx, bson.M{"org_id": orgId})
	return err
}

// DeleteMany — массовое удаление таргетов по списку id
func (m *OrgTargetModel) DeleteMany(ctx context.Context, ids []string) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	oids := make([]primitive.ObjectID, 0, len(ids))
	for _, id := range ids {
		oid, err := primitive.ObjectIDFromHex(id)
		if err == nil {
			oids = append(oids, oid)
		}
	}
	if len(oids) == 0 {
		return 0, nil
	}
	res, err := m.coll.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": oids}})
	if err != nil {
		return 0, err
	}
	return int(res.DeletedCount), nil
}

// Update — обновляет поля таргета
func (m *OrgTargetModel) Update(ctx context.Context, id string, update bson.M) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	update["update_time"] = time.Now()
	_, err = m.coll.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": update})
	return err
}

// CountByOrg — количество таргетов в организации
func (m *OrgTargetModel) CountByOrg(ctx context.Context, orgId string) (int64, error) {
	return m.coll.CountDocuments(ctx, bson.M{"org_id": orgId})
}

// AllEnabled — все включенные таргеты (используется для построения matcher)
func (m *OrgTargetModel) AllEnabled(ctx context.Context) ([]OrgTarget, error) {
	cursor, err := m.coll.Find(ctx, bson.M{"enabled": true})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var docs []OrgTarget
	if err = cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}

// =====================================================================
// OrgTargetMatcher — utility для определения org_id по host/IP
// =====================================================================

// OrgTargetMatcher позволяет быстро определить, к какой организации
// принадлежит host (IP или домен), по списку таргетов.
type OrgTargetMatcher struct {
	// exactDomain[lowercase] -> orgId
	exactDomain map[string]string
	// exactIP[ip] -> orgId
	exactIP map[string]string
	// wildcards: list of (suffix без *., orgId)
	wildcards []wildcardEntry
	// CIDR-сети
	cidrs []cidrEntry
}

type wildcardEntry struct {
	suffix string // например "example.com" (без "*.")
	orgId  string
}

type cidrEntry struct {
	network *net.IPNet
	orgId   string
}

// NewOrgTargetMatcher — создаёт matcher из списка таргетов
func NewOrgTargetMatcher(targets []OrgTarget) *OrgTargetMatcher {
	mr := &OrgTargetMatcher{
		exactDomain: make(map[string]string),
		exactIP:     make(map[string]string),
		wildcards:   make([]wildcardEntry, 0),
		cidrs:       make([]cidrEntry, 0),
	}
	for _, t := range targets {
		if !t.Enabled {
			continue
		}
		v := strings.TrimSpace(t.Value)
		if v == "" {
			continue
		}
		switch t.Type {
		case OrgTargetTypeIP:
			mr.exactIP[v] = t.OrgId
		case OrgTargetTypeCIDR:
			if _, network, err := net.ParseCIDR(v); err == nil {
				mr.cidrs = append(mr.cidrs, cidrEntry{network: network, orgId: t.OrgId})
			}
		case OrgTargetTypeDomain:
			mr.exactDomain[strings.ToLower(v)] = t.OrgId
		case OrgTargetTypeWildcard:
			suffix := strings.ToLower(strings.TrimPrefix(v, "*."))
			if suffix != "" {
				mr.wildcards = append(mr.wildcards, wildcardEntry{suffix: suffix, orgId: t.OrgId})
			}
		}
	}
	return mr
}

// Match — определяет orgId для host. host может быть IP или доменом.
// Возвращает пустую строку если совпадений нет.
// Приоритет: точный IP > CIDR > точный домен > wildcard.
func (m *OrgTargetMatcher) Match(host string) string {
	if m == nil {
		return ""
	}
	h := strings.TrimSpace(host)
	if h == "" {
		return ""
	}
	// убрать порт если есть (host:port)
	if strings.Contains(h, ":") && !strings.Contains(h, "[") {
		if hostPart, _, err := net.SplitHostPort(h); err == nil {
			h = hostPart
		}
	}
	h = strings.ToLower(h)

	// IP-адрес
	if ip := net.ParseIP(h); ip != nil {
		if orgId, ok := m.exactIP[h]; ok {
			return orgId
		}
		for _, c := range m.cidrs {
			if c.network.Contains(ip) {
				return c.orgId
			}
		}
		return ""
	}

	// Домен — точное совпадение
	if orgId, ok := m.exactDomain[h]; ok {
		return orgId
	}
	// Wildcard: example.com и *.example.com совпадает с foo.example.com
	for _, w := range m.wildcards {
		if h == w.suffix || strings.HasSuffix(h, "."+w.suffix) {
			return w.orgId
		}
	}
	return ""
}
