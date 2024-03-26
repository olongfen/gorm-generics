package gormgenerics

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BasicRepo 基本的存储库
type BasicRepo[T any] interface {
	// Create 创建一条数据
	Create(ctx context.Context, ent *T) error
	// Creates 批量创建
	Creates(ctx context.Context, ent []*T) error
	// Update 通过id更新数据
	Update(ctx context.Context, id uint, ent *T) error
	// UpdateColumns 通过id更新部分字段数据
	UpdateColumns(ctx context.Context, id uint, data any) error
	// UpdateColumnsBy 通过条件表达式更新数据
	UpdateColumnsBy(ctx context.Context, expression []clause.Expression, data any) error
	// Find 查询
	Find(ctx context.Context, expression ...clause.Expression) ([]*T, error)
	// FindOneBy 通过条件表达式查询一条数据
	FindOneBy(ctx context.Context, expression []clause.Expression) (*T, error)
	// FindOne 通过id查询一条数据
	FindOne(ctx context.Context, id uint) (*T, error)
	// DeleteOne 通过id删除一条数据
	DeleteOne(ctx context.Context, id uint) error
	// DeleteBy 通过条件删除一条或者多条数据
	DeleteBy(ctx context.Context, expression []clause.Expression) error
	// Count 进行统计数据
	Count(ctx context.Context, expression ...clause.Expression) (count int64, err error)
	// ExecTX 开启事务
	ExecTX(ctx context.Context, fc func(ctx context.Context) error) error
	// Database 获取db
	Database() Database
	// Model 获取模型
	Model() *T
}

type Database interface {
	DB(ctx context.Context) *gorm.DB
	Close() error
	TranslateGormError(ctx context.Context, db *gorm.DB) error
	ExecTx(ctx context.Context, fc func(context.Context) error) error
}

// Limit 数据库数量查询限制
type Limit struct {
	All      bool // true获取全部条目
	PageSize uint // 每页数量
	PageNum  uint // 页数
	Count    bool // true获取总数
}
