package sugar

import (
	"context"

	"github.com/Kirby980/go-es/builder"
	"github.com/Kirby980/go-es/client"
)

// Sugar 对 builder 操作的语法糖封装
type Sugar struct {
	client *client.Client
}

// New 创建 Sugar 实例
func New(c *client.Client) *Sugar {
	return &Sugar{client: c}
}

// AutoMigrate 自动迁移索引（GORM 风格）
// 索引名通过以下方式确定：
// 1. 如果 model 实现了 IndexName() string 方法，使用该方法返回值
// 2. 否则使用结构体名称转 snake_case + s
// 3. 由于 ES 不支持删除字段，只能添加新字段
func (s *Sugar) AutoMigrate(models ...any) error {
	for _, model := range models {
		ib := builder.NewIndexBuilder(s.client, "").AutoMigrate(model)
		ctx := context.Background()
		exists, err := ib.Exists(ctx)
		if err != nil {
			return err
		}
		if exists {
			if err := ib.PutMapping(ctx); err != nil {
				return err
			}
		} else {
			if err := ib.Create(ctx); err != nil {
				return err
			}
		}
	}
	return nil
}

// Create 创建文档（GORM 风格）
// model 必须是结构体指针，索引名自动从 IndexName() 方法或结构体名推断
func (s *Sugar) Create(ctx context.Context, model any) (*builder.DocumentResponse, error) {
	return builder.NewDocumentBuilder(s.client, "").Model(model).Do(ctx)
}

// CreateWithID 创建文档并指定 ID
func (s *Sugar) CreateWithID(ctx context.Context, id string, model any) (*builder.DocumentResponse, error) {
	return builder.NewDocumentBuilder(s.client, "").Model(model).ID(id).Do(ctx)
}

// Update 更新文档（部分更新）
func (s *Sugar) Update(ctx context.Context, id string, model any) (*builder.DocumentResponse, error) {
	return builder.NewDocumentBuilder(s.client, "").Model(model).ID(id).Update(ctx)
}

// Upsert 更新或插入文档
func (s *Sugar) Upsert(ctx context.Context, id string, model any) (*builder.DocumentResponse, error) {
	return builder.NewDocumentBuilder(s.client, "").Model(model).ID(id).Upsert(ctx)
}

// Delete 删除文档
func (s *Sugar) Delete(ctx context.Context, index, id string) (*builder.DocumentResponse, error) {
	return builder.NewDocumentBuilder(s.client, index).ID(id).Delete(ctx)
}

// Get 获取单个文档
func (s *Sugar) Get(ctx context.Context, index, id string) (*builder.GetResponse, error) {
	return builder.NewDocumentBuilder(s.client, index).ID(id).Get(ctx)
}

// Find 创建搜索构建器，用于链式查询
// 示例：sugar.New(c).Find("index").Term("field", "value").Size(10).Do(ctx)
func (s *Sugar) Find(index string) *builder.SearchBuilder {
	return builder.NewSearchBuilder(s.client, index)
}
