package http

import (
	"net/http"
	"reflect"
	"testing"
)

// TestEmptyCallOptions 测试空调用选项
func TestEmptyCallOptions(t *testing.T) {
	e := EmptyCallOption{}
	// 测试 before 方法是否返回 nil
	if e.before(&callInfo{}) != nil {
		t.Error("EmptyCallOption should be ignored")
	}
	// 测试 after 方法是否正常工作
	e.after(&callInfo{}, &csAttempt{})
}

// TestContentType 测试内容类型
func TestContentType(t *testing.T) {
	// 测试 ContentType 函数是否返回正确的内容类型
	if !reflect.DeepEqual(ContentType("aaa").(ContentTypeCallOption).ContentType, "aaa") {
		t.Errorf("want: %v,got: %v", "aaa", ContentType("aaa").(ContentTypeCallOption).ContentType)
	}
}

// TestContentTypeCallOption_before 测试内容类型调用选项的 before 方法
func TestContentTypeCallOption_before(t *testing.T) {
	c := &callInfo{}
	// 测试 before 方法是否正确设置内容类型
	err := ContentType("aaa").before(c)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual("aaa", c.contentType) {
		t.Errorf("want: %v, got: %v", "aaa", c.contentType)
	}
}

// TestDefaultCallInfo 测试默认调用信息
func TestDefaultCallInfo(t *testing.T) {
	path := "hi"
	// 测试 defaultCallInfo 函数是否返回正确的调用信息
	rv := defaultCallInfo(path)
	if !reflect.DeepEqual(path, rv.pathTemplate) {
		t.Errorf("expect %v, got %v", path, rv.pathTemplate)
	}
	if !reflect.DeepEqual(path, rv.operation) {
		t.Errorf("expect %v, got %v", path, rv.operation)
	}
	if !reflect.DeepEqual("application/json", rv.contentType) {
		t.Errorf("expect %v, got %v", "application/json", rv.contentType)
	}
}

// TestOperation 测试操作
func TestOperation(t *testing.T) {
	// 测试 Operation 函数是否返回正确的操作
	if !reflect.DeepEqual("aaa", Operation("aaa").(OperationCallOption).Operation) {
		t.Errorf("want: %v,got: %v", "aaa", Operation("aaa").(OperationCallOption).Operation)
	}
}

// TestOperationCallOption_before 测试操作调用选项的 before 方法
func TestOperationCallOption_before(t *testing.T) {
	c := &callInfo{}
	// 测试 before 方法是否正确设置操作
	err := Operation("aaa").before(c)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual("aaa", c.operation) {
		t.Errorf("want: %v, got: %v", "aaa", c.operation)
	}
}

// TestPathTemplate 测试路径模板
func TestPathTemplate(t *testing.T) {
	// 测试 PathTemplate 函数是否返回正确的路径模板
	if !reflect.DeepEqual("aaa", PathTemplate("aaa").(PathTemplateCallOption).Pattern) {
		t.Errorf("want: %v,got: %v", "aaa", PathTemplate("aaa").(PathTemplateCallOption).Pattern)
	}
}

// TestPathTemplateCallOption_before 测试路径模板调用选项的 before 方法
func TestPathTemplateCallOption_before(t *testing.T) {
	c := &callInfo{}
	// 测试 before 方法是否正确设置路径模板
	err := PathTemplate("aaa").before(c)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual("aaa", c.pathTemplate) {
		t.Errorf("want: %v, got: %v", "aaa", c.pathTemplate)
	}
}

// TestHeader 测试头部
func TestHeader(t *testing.T) {
	h := http.Header{"A": []string{"123"}}
	// 测试 Header 函数是否返回正确的头部
	if !reflect.DeepEqual(Header(&h).(HeaderCallOption).header.Get("A"), "123") {
		t.Errorf("want: %v,got: %v", "123", Header(&h).(HeaderCallOption).header.Get("A"))
	}
}

// TestHeaderCallOption_before 测试头部调用选项的 before 方法
func TestHeaderCallOption_before(t *testing.T) {
	h := http.Header{"A": []string{"123"}}
	c := &callInfo{}
	o := Header(&h)
	// 测试 before 方法是否正确设置头部
	err := o.before(c)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(&h, c.headerCarrier) {
		t.Errorf("want: %v,got: %v", &h, o.(HeaderCallOption).header)
	}
}

// TestHeaderCallOption_after 测试头部调用选项的 after 方法
func TestHeaderCallOption_after(t *testing.T) {
	h := http.Header{"A": []string{"123"}}
	c := &callInfo{}
	cs := &csAttempt{res: &http.Response{Header: h}}
	o := Header(&h)
	// 测试 after 方法是否正确设置头部
	o.after(c, cs)
	if !reflect.DeepEqual(&h, o.(HeaderCallOption).header) {
		t.Errorf("want: %v,got: %v", &h, o.(HeaderCallOption).header)
	}
}
