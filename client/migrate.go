package client

import (
	"context"

	"github.com/Kirby980/go-es/builder"
)

// AutoMigrate 自动迁移索引（GORM 风格）
// 索引名通过以下方式确定：
// 1. 如果 model 实现了 IndexName() string 方法，使用该方法返回值
// 2. 否则使用结构体名称转 snake_cases
// 3.由于es不支持删除索引，所以只能添加索引字段
func (c *Client) AutoMigrate(models ...any) error {
	for _, model := range models {
		// 创建 IndexBuilder，索引名传空字符串，让 AutoMigrate 自动推断
		ib := builder.NewIndexBuilder(c, "").AutoMigrate(model)
		ctx := context.Background()
		// 检查索引是否存在
		exists, err := ib.Exists(ctx)
		if err != nil {
			return err
		}

		if exists {
			// 索引已存在，更新映射（添加新字段）
			if err := ib.PutMapping(ctx); err != nil {
				return err
			}
		} else {
			// 索引不存在，创建
			if err := ib.Create(ctx); err != nil {
				return err
			}
		}
	}
	return nil
}

// Create 创建文档（GORM 风格）
// model 必须是结构体指针，索引名自动从 IndexName() 方法或结构体名推断
func (c *Client) Create(ctx context.Context, model any) (*builder.DocumentResponse, error) {
	return builder.NewDocumentBuilder(c, "").Model(model).Do(ctx)
}

// CreateWithID 创建文档并指定 ID
func (c *Client) CreateWithID(ctx context.Context, id string, model any) (*builder.DocumentResponse, error) {
	return builder.NewDocumentBuilder(c, "").Model(model).ID(id).Do(ctx)
}

// Update 更新文档
func (c *Client) Update(ctx context.Context, id string, model any) (*builder.DocumentResponse, error) {
	return builder.NewDocumentBuilder(c, "").Model(model).ID(id).Update(ctx)
}

// Upsert 更新或插入文档
func (c *Client) Upsert(ctx context.Context, id string, model any) (*builder.DocumentResponse, error) {
	return builder.NewDocumentBuilder(c, "").Model(model).ID(id).Upsert(ctx)
}

// Delete 删除文档
func (c *Client) Delete(ctx context.Context, index, id string) (*builder.DocumentResponse, error) {
	return builder.NewDocumentBuilder(c, index).ID(id).Delete(ctx)
}

// Get 获取单个文档
func (c *Client) Get(ctx context.Context, index, id string) (*builder.GetResponse, error) {
	return builder.NewDocumentBuilder(c, index).ID(id).Get(ctx)
}

// Find 搜索文档并扫描到结构体切片
// dest 必须是指向切片的指针，如 *[]Article
// 返回 SearchBuilder 以便添加更多查询条件
func (c *Client) Find(index string) *builder.SearchBuilder {
	return builder.NewSearchBuilder(c, index)
}
