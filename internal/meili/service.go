package meili

import (
	"encoding/json"
	"fmt"
)

const indexPath = "/indexes"

// IndexCreateBody 创建索引 body
type IndexCreateBody struct {
	UID        string `json:"uid"`
	PrimaryKey string `json:"primaryKey,omitempty"`
}

// Indexes 索引列表/单索引响应由 Meilisearch 返回，直接透传 JSON

// IndexList 获取所有索引
func (c *Client) IndexList() ([]byte, int, error) {
	return c.Get(indexPath)
}

// IndexCreate 创建索引
func (c *Client) IndexCreate(uid, primaryKey string) ([]byte, int, error) {
	return c.Post(indexPath, IndexCreateBody{UID: uid, PrimaryKey: primaryKey})
}

// IndexGet 获取单个索引
func (c *Client) IndexGet(uid string) ([]byte, int, error) {
	return c.Get(indexPath + "/" + uid)
}

// IndexUpdate 更新索引设置（settings）
func (c *Client) IndexUpdate(uid string, body map[string]interface{}) ([]byte, int, error) {
	// Meilisearch 更新设置使用 PATCH /indexes/:uid/settings
	return c.Patch(indexPath+"/"+uid+"/settings", body)
}

// IndexDelete 删除索引
func (c *Client) IndexDelete(uid string) ([]byte, int, error) {
	return c.Delete(indexPath + "/" + uid)
}

// DocAdd 添加/替换文档，documents 为对象数组
func (c *Client) DocAdd(indexUID string, documents interface{}) ([]byte, int, error) {
	return c.Put(indexPath+"/"+indexUID+"/documents", documents)
}

// DocList 获取文档列表，limit/offset 传 0 表示不传该参数
func (c *Client) DocList(indexUID string, limit, offset int) ([]byte, int, error) {
	path := indexPath + "/" + indexUID + "/documents"
	if limit > 0 || offset > 0 {
		path += "?"
		if limit > 0 {
			path += fmt.Sprintf("limit=%d", limit)
		}
		if offset > 0 {
			if limit > 0 {
				path += "&"
			}
			path += fmt.Sprintf("offset=%d", offset)
		}
	}
	return c.Get(path)
}

// DocGet 获取单条文档
func (c *Client) DocGet(indexUID, documentID string) ([]byte, int, error) {
	return c.Get(indexPath + "/" + indexUID + "/documents/" + documentID)
}

// DocDeleteOne 删除单条文档
func (c *Client) DocDeleteOne(indexUID, documentID string) ([]byte, int, error) {
	return c.Delete(indexPath + "/" + indexUID + "/documents/" + documentID)
}

// DocDeleteBatchBody 批量删除 body
type DocDeleteBatchBody struct {
	Ids []string `json:"ids"`
}

// DocDeleteBatch 批量删除文档
func (c *Client) DocDeleteBatch(indexUID string, ids []string) ([]byte, int, error) {
	return c.Post(indexPath+"/"+indexUID+"/documents/delete-batch", DocDeleteBatchBody{Ids: ids})
}

// DocDeleteAll 删除索引下所有文档
func (c *Client) DocDeleteAll(indexUID string) ([]byte, int, error) {
	return c.Delete(indexPath + "/" + indexUID + "/documents")
}

// DocUpdate 更新文档（部分字段），body 为对象或数组，会转 JSON
func (c *Client) DocUpdate(indexUID string, documents interface{}) ([]byte, int, error) {
	b, err := json.Marshal(documents)
	if err != nil {
		return nil, 0, err
	}
	return c.Put(indexPath+"/"+indexUID+"/documents", json.RawMessage(b))
}

// SearchMeilisearch 在 /indexes/:uid/search 做查询（POST）
func (c *Client) Search(indexUID string, body any) ([]byte, int, error) {
	return c.Post(indexPath+"/"+indexUID+"/search", body)
}
