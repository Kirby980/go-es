package builder

import "encoding/json"

// NestedObject 嵌套对象构建器（用于构建嵌套字段）
type NestedObject struct {
	data map[string]interface{}
}

// Set 设置嵌套对象的字段
func (o *NestedObject) Set(key string, value interface{}) *NestedObject {
	o.data[key] = value
	return o
}

// SetObject 设置嵌套对象的子对象
func (o *NestedObject) SetObject(key string, builder func(*NestedObject)) *NestedObject {
	nested := &NestedObject{data: make(map[string]interface{})}
	builder(nested)
	o.data[key] = nested.data
	return o
}

// SetArray 设置数组字段
func (o *NestedObject) SetArray(key string, values ...interface{}) *NestedObject {
	o.data[key] = values
	return o
}

// SetObjectArray 设置对象数组
func (o *NestedObject) SetObjectArray(key string, builders ...func(*NestedObject)) *NestedObject {
	arr := make([]map[string]interface{}, len(builders))
	for i, builder := range builders {
		nested := &NestedObject{data: make(map[string]interface{})}
		builder(nested)
		arr[i] = nested.data
	}
	o.data[key] = arr
	return o
}

// SetFromStruct 从结构体设置字段
func (o *NestedObject) SetFromStruct(data interface{}) *NestedObject {
	jsonData, _ := json.Marshal(data)
	json.Unmarshal(jsonData, &o.data)
	return o
}
