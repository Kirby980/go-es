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
func (c *Client) AutoMigrate(models ...interface{}) error {
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
