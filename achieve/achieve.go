package achieve

import (
	"context"
	"errors"
	gorm_generics "gorm-generics"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// basicRepo 基础存储库实现
type basicRepo[T any] struct {
	database gorm_generics.Database
}

var TranslateError func(ctx context.Context, db *gorm.DB) error

func init() {
	TranslateError = func(ctx context.Context, db *gorm.DB) error {
		return nil
	}
}

// NewBasicRepository 新建基础存储库
func NewBasicRepository[T any](database gorm_generics.Database) gorm_generics.BasicRepo[T] {
	return &basicRepo[T]{database}
}

// Model 返回模型实体
func (b *basicRepo[T]) Model() *T {
	var model T
	return &model
}

// Database 返回db
func (b *basicRepo[T]) Database() gorm_generics.Database {
	return b.database
}

// Create 创建一条记录
func (b *basicRepo[T]) Create(ctx context.Context, ent *T) error {
	db := b.database.DB(ctx).Create(ent)
	if err := db.Error; err != nil {
		return TranslateError(ctx, db)
	}
	return nil
}

// Creates 创建一条或者多条记录
func (b *basicRepo[T]) Creates(ctx context.Context, ent []*T) error {
	db := b.database.DB(ctx).Create(ent)
	if err := db.Error; err != nil {
		return TranslateError(ctx, db)
	}
	return nil
}

// processExpression 处理条件
func processExpression(db *gorm.DB, conds []clause.Expression) *gorm.DB {
	for _, v := range conds {
		val, ok := v.(clause.OrderBy)
		if ok {
			for _, order := range val.Columns {
				db = db.Order(order)
			}
			continue
		}
		db = db.Where(v)
	}
	return db
}

// Find 查询
func (b *basicRepo[T]) Find(ctx context.Context, limit *gorm_generics.Limit, conds ...clause.Expression) ([]*T, int64, error) {
	db := b.database.DB(ctx).Model(b.Model())
	db = processExpression(db, conds)
	var (
		count int64
	)
	// 如果需要全部数据
	if limit.Count {
		if err := db.Count(&count).Error; err != nil {
			err = TranslateError(ctx, db)
			return nil, 0, err
		}
	}
	var (
		data []*T
	)
	if !limit.All {
		switch {
		case limit.PageSize > 0 && limit.PageNum > 0:
			db = db.Offset(int((limit.PageNum - 1) * limit.PageSize)).Limit(int(limit.PageSize))
		case limit.PageSize > 0 && limit.PageNum == 0:
			db = db.Limit(int(limit.PageSize))
		default:
			limit.PageSize = 10
			limit.PageNum = 1
			db = db.Offset(int((limit.PageNum - 1) * limit.PageSize)).Limit(int(limit.PageSize))
		}
	}

	if err := db.Find(&data).Error; err != nil {
		err = TranslateError(ctx, db)
		return nil, 0, err
	}
	return data, count, nil
}

// FindOne 查询一条数据
func (b *basicRepo[T]) FindOne(ctx context.Context, id uint) (*T, error) {
	var (
		data *T
	)
	db := b.database.DB(ctx).Model(b.Model()).Where("id = ?", id).First(&data)
	if err := db.Error; err != nil {
		err = TranslateError(ctx, db)
		return nil, err
	}
	return data, nil
}

// FindOneBy 多条件查询一条数据
func (b *basicRepo[T]) FindOneBy(ctx context.Context, conds []clause.Expression) (*T, error) {
	var (
		data *T
	)
	db := b.database.DB(ctx).Model(b.Model())
	db = processExpression(db, conds)
	if err := db.First(&data).Error; err != nil {
		err = TranslateError(ctx, db)
		return nil, err
	}
	return data, nil
}

// Count 统计
func (b *basicRepo[T]) Count(ctx context.Context, conds []clause.Expression) (count int64, err error) {
	db := b.database.DB(ctx).Model(b.Model())
	db = processExpression(db, conds)
	// 如果需要全部数据
	if err = db.Count(&count).Error; err != nil {
		err = TranslateError(ctx, db)
		return
	}
	return
}

// DeleteOne 删除一条数据
func (b *basicRepo[T]) DeleteOne(ctx context.Context, id uint) error {
	var (
		model T
	)
	db := b.database.DB(ctx).Where("id = ?", id).Delete(&model)
	if err := db.Error; err != nil {
		return TranslateError(ctx, db)
	}
	return nil
}

// DeleteBy 通过条件删除一条或者多条数据
func (b *basicRepo[T]) DeleteBy(ctx context.Context, conds []clause.Expression) error {
	var (
		model T
	)
	if len(conds) == 0 {
		return errors.New("delete condition is empty")
	}
	db := b.database.DB(ctx).Model(b.Model())
	db = processExpression(db, conds)
	db = db.Delete(&model)
	if err := db.Error; err != nil {
		return TranslateError(ctx, db)
	}
	return nil
}

// Update 更新一条数据
func (b *basicRepo[T]) Update(ctx context.Context, id uint, ent *T) error {
	db := b.database.DB(ctx).Model(b.Model()).Where("id = ?", id).Session(&gorm.Session{FullSaveAssociations: true}).Updates(ent)
	if err := db.Error; err != nil {
		err = TranslateError(ctx, db)
		return err
	}
	return nil
}

// UpdateColumns 通过id更新部分字段数据
func (b *basicRepo[T]) UpdateColumns(ctx context.Context, id uint, data any) error {
	db := b.database.DB(ctx).Model(b.Model()).Where("id = ?", id).Save(data)
	if err := db.Error; err != nil {
		err = TranslateError(ctx, db)
		return err
	}
	return nil
}

// UpdateColumnsBy 通过条件表达式更新数据
func (b *basicRepo[T]) UpdateColumnsBy(ctx context.Context, conds []clause.Expression, data any) error {
	if len(conds) == 0 {
		return errors.New("the update condition is empty")
	}
	db := b.database.DB(ctx).Model(b.Model())
	db = processExpression(db, conds)
	db = db.Updates(data)
	if err := db.Error; err != nil {
		err = TranslateError(ctx, db)
		return err
	}
	return nil
}

func (b *basicRepo[T]) ExecTX(ctx context.Context, fc func(ctx context.Context) error) error {
	return b.database.ExecTx(ctx, fc)
}
