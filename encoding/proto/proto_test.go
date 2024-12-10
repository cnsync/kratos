package proto

import (
	"reflect"
	"testing"

	testData "github.com/cnsync/kratos/internal/testdata/encoding"
)

// TestName 测试 codec 的 Name 方法是否返回正确的名称
func TestName(t *testing.T) {
	c := new(codec)
	// 检查 Name 方法返回的名称是否与预期一致
	if !reflect.DeepEqual(c.Name(), "proto") {
		t.Errorf("no expect float_key value: %v, but got: %v", c.Name(), "proto")
	}
}

// TestCodec 测试 codec 的 Marshal 和 Unmarshal 方法是否能正确处理 TestModel 结构体
func TestCodec(t *testing.T) {
	c := new(codec)

	model := testData.TestModel{
		Id:    1,
		Name:  "kratos",
		Hobby: []string{"study", "eat", "play"},
	}

	// 测试 Marshal 方法是否能正确序列化 TestModel 结构体
	m, err := c.Marshal(&model)
	if err != nil {
		t.Errorf("Marshal() should be nil, but got %s", err)
	}

	var res testData.TestModel

	// 测试 Unmarshal 方法是否能正确反序列化字节切片到 TestModel 结构体
	err = c.Unmarshal(m, &res)
	if err != nil {
		t.Errorf("Unmarshal() should be nil, but got %s", err)
	}
	// 检查反序列化后的 TestModel 结构体的各个字段是否与原始结构体一致
	if !reflect.DeepEqual(res.Id, model.Id) {
		t.Errorf("ID should be %d, but got %d", res.Id, model.Id)
	}
	if !reflect.DeepEqual(res.Name, model.Name) {
		t.Errorf("Name should be %s, but got %s", res.Name, model.Name)
	}
	if !reflect.DeepEqual(res.Hobby, model.Hobby) {
		t.Errorf("Hobby should be %s, but got %s", res.Hobby, model.Hobby)
	}
}

// TestCodec2 测试 codec 的 Unmarshal 方法是否能正确处理指针类型的目标参数
func TestCodec2(t *testing.T) {
	c := new(codec)

	model := testData.TestModel{
		Id:    1,
		Name:  "kratos",
		Hobby: []string{"study", "eat", "play"},
	}

	m, err := c.Marshal(&model)
	if err != nil {
		t.Errorf("Marshal() should be nil, but got %s", err)
	}

	var res testData.TestModel
	rp := &res

	// 测试 Unmarshal 方法是否能正确反序列化字节切片到指针类型的 TestModel 结构体
	err = c.Unmarshal(m, &rp)
	if err != nil {
		t.Errorf("Unmarshal() should be nil, but got %s", err)
	}
	// 检查反序列化后的 TestModel 结构体的各个字段是否与原始结构体一致
	if !reflect.DeepEqual(res.Id, model.Id) {
		t.Errorf("ID should be %d, but got %d", res.Id, model.Id)
	}
	if !reflect.DeepEqual(res.Name, model.Name) {
		t.Errorf("Name should be %s, but got %s", res.Name, model.Name)
	}
	if !reflect.DeepEqual(res.Hobby, model.Hobby) {
		t.Errorf("Hobby should be %s, but got %s", res.Hobby, model.Hobby)
	}
}

// Test_getProtoMessage 测试 getProtoMessage 函数是否能正确处理不同类型的参数
func Test_getProtoMessage(t *testing.T) {
	p := &testData.TestModel{Id: 1}
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// 测试传入的参数是 TestModel 结构体指针时，函数是否能正确返回 proto.Message 接口的实现
		{name: "test1", args: args{v: &testData.TestModel{}}, wantErr: false},
		// 测试传入的参数是 TestModel 结构体时，函数是否能正确返回 proto.Message 接口的实现
		{name: "test2", args: args{v: testData.TestModel{}}, wantErr: true},
		// 测试传入的参数是 TestModel 结构体指针的指针时，函数是否能正确返回 proto.Message 接口的实现
		{name: "test3", args: args{v: &p}, wantErr: false},
		// 测试传入的参数是 int 类型时，函数是否能正确返回 proto.Message 接口的实现
		{name: "test4", args: args{v: 1}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := getProtoMessage(tt.args.v)
			// 检查函数返回的错误是否与预期一致
			if (err != nil) != tt.wantErr {
				t.Errorf("getProtoMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
