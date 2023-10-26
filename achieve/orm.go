package achieve

import (
	"context"
	gorm_generics "github.com/olongfen/gorm-generics"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GormSpanKey 包内静态变量
const GormSpanKey = "__gorm_span"

const (
	// CallBackBeforeName sql执行之前tag
	CallBackBeforeName = "opentracing:before"
	// CallBackAfterName sql执行之后tag
	CallBackAfterName = "opentracing:after"
)

type DB struct {
	db           *gorm.DB
	translateErr func(ctx context.Context, db *gorm.DB) error
}

func (d *DB) TranslateGormError(ctx context.Context, db *gorm.DB) error {
	if d.translateErr != nil {
		return d.translateErr(ctx, db)
	}
	return db.Error
}

// NewDataBase new database
func NewDataBase(driver DriverName, dsn string, opts ...Option) (gorm_generics.Database, error) {
	var (
		option = options{}
	)
	for _, o := range opts {
		o.apply(&option)
	}
	db, err := DBConnect(driver, dsn, option)
	if err != nil {
		return nil, err
	}
	database := &DB{db, option.translateError}
	return database, nil
}

// contextTxKey 事务上下文 key
type contextTxKey struct{}

// ExecTx 执行事务
func (d *DB) ExecTx(ctx context.Context, fc func(context.Context) error) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ctx = context.WithValue(ctx, contextTxKey{}, tx)
		return fc(ctx)
	})
}

// DB 获取db
func (d *DB) DB(ctx context.Context) *gorm.DB {
	tx, ok := ctx.Value(contextTxKey{}).(*gorm.DB)
	if ok {
		return tx
	}
	return d.db.WithContext(ctx)
}

// Close 关闭db连接
func (d *DB) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// OpentracingPlugin 追踪插件
type OpentracingPlugin struct {
}

var _ gorm.Plugin = (*OpentracingPlugin)(nil)

func (op *OpentracingPlugin) Name() string {
	return "opentracingPlugin"
}

func (op *OpentracingPlugin) Initialize(db *gorm.DB) error {
	// 开始前 - 并不是都用相同的方法，可以自己自定义
	if err := db.Callback().Create().Before("gorm:before_create").Register(CallBackBeforeName, before); err != nil {
		return err
	}
	if err := db.Callback().Query().Before("gorm:query").Register(CallBackBeforeName, before); err != nil {
		return err
	}
	if err := db.Callback().Delete().Before("gorm:before_delete").Register(CallBackBeforeName, before); err != nil {
		return err
	}
	if err := db.Callback().Update().Before("gorm:setup_reflect_value").Register(CallBackBeforeName, before); err != nil {
		return err
	}
	if err := db.Callback().Row().Before("gorm:row").Register(CallBackBeforeName, before); err != nil {
		return err
	}
	if err := db.Callback().Raw().Before("gorm:raw").Register(CallBackBeforeName, before); err != nil {
		return err
	}

	// 结束后 - 并不是都用相同的方法，可以自己自定义
	if err := db.Callback().Create().After("gorm:after_create").Register(CallBackAfterName, after); err != nil {
		return err
	}
	if err := db.Callback().Query().After("gorm:after_query").Register(CallBackAfterName, after); err != nil {
		return err
	}
	if err := db.Callback().Delete().After("gorm:after_delete").Register(CallBackAfterName, after); err != nil {
		return err
	}
	if err := db.Callback().Update().After("gorm:after_update").Register(CallBackAfterName, after); err != nil {
		return err
	}
	if err := db.Callback().Row().After("gorm:row").Register(CallBackAfterName, after); err != nil {
		return err
	}
	if err := db.Callback().Raw().After("gorm:raw").Register(CallBackAfterName, after); err != nil {
		return err
	}
	return nil
}

func before(db *gorm.DB) {
	tr := otel.Tracer("gorm-before")
	_, span := tr.Start(db.Statement.Context, "gorm-before")
	// 利用db实例去传递span
	db.InstanceSet(GormSpanKey, span)
}

func after(db *gorm.DB) {
	_span, exist := db.InstanceGet(GormSpanKey)
	if !exist {
		return
	}
	// 断言类型
	span, ok := _span.(trace.Span)
	if !ok {
		return
	}

	defer span.End()

	if db.Error != nil {
		span.SetAttributes(attribute.Key("gorm_err").String(db.Error.Error()))
	}
	span.SetAttributes(attribute.Key("sql").String(db.Dialector.Explain(db.Statement.SQL.String(), db.Statement.Vars...)))
}

// DBWhereExpression process field symbol expression
func DBWhereExpression(column string, value any, symbol string) clause.Expression {
	switch symbol {
	case ">":
		return clause.Gt{Column: column, Value: value}
	case ">=":
		return clause.Gte{Column: column, Value: value}
	case "<":
		return clause.Lt{Column: column, Value: value}
	case "<=":
		return clause.Lte{Column: column, Value: value}
	case "like":
		return clause.Like{Column: column, Value: value}
	case "ilike":
		return ILike{Column: column, Value: value}
	case "in":
		return clause.IN{Column: column, Values: gormParseIn(value)}
	case "expr":
		return clause.Expr{
			SQL: column,
		}
	default:
		return clause.Eq{Column: column, Value: value}
	}
}

// gormParseIn gorm in 值解析
func gormParseIn(value any) []any {
	var (
		ret []any
	)
	arrString, ok := value.([]string)
	if ok {
		ret = make([]any, 0, len(arrString))
		for v := range arrString {
			ret = append(ret, v)
		}
		return ret
	}

	arrInt, ok := value.([]int)
	if ok {
		ret = make([]any, 0, len(arrInt))
		for v := range arrString {
			ret = append(ret, v)
		}
		return ret
	}

	arrInt8, ok := value.([]int8)
	if ok {
		ret = make([]any, 0, len(arrInt8))
		for _, v := range arrInt8 {
			ret = append(ret, v)
		}
		return ret
	}

	arrInt16, ok := value.([]int16)
	if ok {
		ret = make([]any, 0, len(arrInt16))
		for _, v := range arrInt16 {
			ret = append(ret, v)
		}
		return ret
	}

	arrInt32, ok := value.([]int32)
	if ok {
		ret = make([]any, 0, len(arrInt32))
		for _, v := range arrInt32 {
			ret = append(ret, v)
		}
	}
	arrInt64, ok := value.([]int32)
	if ok {
		ret = make([]any, 0, len(arrInt64))
		for _, v := range arrInt64 {
			ret = append(ret, v)
		}
		return ret
	}

	arrTime, ok := value.([]*time.Time)
	if ok {
		ret = make([]any, 0, len(arrTime))
		for _, v := range arrTime {
			ret = append(ret, v)
		}
		return ret
	}
	arrBool, ok := value.([]bool)
	if ok {
		ret = make([]any, 0, len(arrBool))
		for v := range arrBool {
			ret = append(ret, v)
		}
		return ret
	}

	arrFloat32, ok := value.([]float32)
	if ok {
		ret = make([]any, 0, len(arrFloat32))
		for v := range arrFloat32 {
			ret = append(ret, v)
		}
		return ret
	}
	arrFloat64, ok := value.([]float32)
	if ok {
		ret = make([]any, 0, len(arrFloat64))
		for v := range arrFloat64 {
			ret = append(ret, v)
		}
		return ret
	}
	//default:
	ret = append(ret, value)
	return ret
}
