package gormgenerics

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// basicRepo 基础存储库实现
type basicRepo[T any] struct {
	database Database
}

// NewBasicRepository 新建基础存储库
func NewBasicRepository[T any](database Database) BasicRepo[T] {
	return &basicRepo[T]{database}
}

// Model 返回模型实体
func (b *basicRepo[T]) Model() *T {
	var model T
	return &model
}

// Database 返回db
func (b *basicRepo[T]) Database() Database {
	return b.database
}

// Create 创建一条记录
func (b *basicRepo[T]) Create(ctx context.Context, ent *T) error {
	db := b.database.DB(ctx).Create(ent)
	if err := db.Error; err != nil {
		return b.database.TranslateGormError(ctx, db)
	}
	return nil
}

// Creates 创建一条或者多条记录
func (b *basicRepo[T]) Creates(ctx context.Context, ent []*T) error {
	db := b.database.DB(ctx).Create(ent)
	if err := db.Error; err != nil {
		return b.database.TranslateGormError(ctx, db)
	}
	return nil
}

// Find 查询
func (b *basicRepo[T]) Find(ctx context.Context, conds ...clause.Expression) ([]*T, error) {
	db := b.database.DB(ctx).Model(b.Model())
	var (
		data []*T
	)
	if err := db.Clauses(conds...).Find(&data).Error; err != nil {
		err = b.database.TranslateGormError(ctx, db)
		return nil, err
	}
	return data, nil
}

// FindOne 查询一条数据
func (b *basicRepo[T]) FindOne(ctx context.Context, id uint) (*T, error) {
	var (
		data *T
	)
	db := b.database.DB(ctx).Model(b.Model()).Where("id = ?", id).First(&data)
	if err := db.Error; err != nil {
		err = b.database.TranslateGormError(ctx, db)
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
	if err := db.Clauses(conds...).First(&data).Error; err != nil {
		err = b.database.TranslateGormError(ctx, db)
		return nil, err
	}
	return data, nil
}

// Count 统计
func (b *basicRepo[T]) Count(ctx context.Context, conds ...clause.Expression) (count int64, err error) {
	db := b.database.DB(ctx).Model(b.Model())
	// 如果需要全部数据
	if err = db.Clauses(conds...).Count(&count).Error; err != nil {
		err = b.database.TranslateGormError(ctx, db)
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
		return b.database.TranslateGormError(ctx, db)
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
	db = db.Clauses(conds...).Delete(&model)
	if err := db.Error; err != nil {
		return b.database.TranslateGormError(ctx, db)
	}
	return nil
}

// Update 更新一条数据
func (b *basicRepo[T]) Update(ctx context.Context, id uint, ent *T) error {
	db := b.database.DB(ctx).Model(b.Model()).Where("id = ?", id).Session(&gorm.Session{FullSaveAssociations: true}).Updates(ent)
	if err := db.Error; err != nil {
		err = b.database.TranslateGormError(ctx, db)
		return err
	}
	return nil
}

// UpdateColumns 通过id更新部分字段数据
func (b *basicRepo[T]) UpdateColumns(ctx context.Context, id uint, data any) error {
	db := b.database.DB(ctx).Model(b.Model()).Where("id = ?", id).Save(data)
	if err := db.Error; err != nil {
		err = b.database.TranslateGormError(ctx, db)
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
	db = db.Clauses(conds...).Updates(data)
	if err := db.Error; err != nil {
		err = b.database.TranslateGormError(ctx, db)
		return err
	}
	return nil
}

func (b *basicRepo[T]) ExecTX(ctx context.Context, fc func(ctx context.Context) error) error {
	return b.database.ExecTx(ctx, fc)
}
