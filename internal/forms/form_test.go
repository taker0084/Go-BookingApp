package forms

import (
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestForm_Valid(t *testing.T){
	r :=httptest.NewRequest("POST","/whatever",nil)
	form := New(r.PostForm)

	isValid := form.Valid()
	if !isValid{
		t.Error("got invalid when should have been valid")
	}
}
func TestForm_Required(t *testing.T){
	r :=httptest.NewRequest("POST", "/whatever", nil)
	form := New(r.PostForm)

	form.Required("a", "b", "c")
	if form.Valid(){
		t.Error("form shows valid when required fields missing")
	}

	postedData := url.Values{}
	postedData.Add("a", "a")
	postedData.Add("b", "b")
	postedData.Add("c", "c")

	r = httptest.NewRequest("POST", "/whatever", nil)

	r.PostForm = postedData
	form = New(r.PostForm)
	form.Required("a", "b", "c")
	if !form.Valid(){
		t.Error("shows does not have required fields when it does")
	}
}

func TestHas(t *testing.T){
	postedData := url.Values{}
	form := New(postedData)

	has := form.Has("whatever")
	if has{
		t.Error("form shows has field when it does not")
	}

	postedData = url.Values{}
	postedData.Add("a", "a")

	form = New(postedData)

	has = form.Has("a")
	if !has{
		t.Error("shows form does not have field when it should")
	}
}

func TestMinLength(t *testing.T){
	postedData := url.Values{}
	form := New(postedData)

	form.MinLength("x", 10)
	if form.Valid(){
		t.Error("form shows min length for non-existent field")
	}

	isError := form.Errors.Get("x")
	if isError == ""{
		t.Error("should have an error but did not get one")
	}

	postedData = url.Values{}
	postedData.Add("x", "some value")

	form = New(postedData)

	form.MinLength("x", 100)
	if form.Valid(){
		t.Error("shows min length of 5 met when data shorter")
	}

	postedData = url.Values{}
	postedData.Add("another_field", "abc123")
	form=New(postedData)

	form.MinLength("another_field", 1)
	if !form.Valid(){
		t.Error("shows min length of 1 is not met when it is")
	}
	isError = form.Errors.Get("another_field")
	if isError != ""{
		t.Error("should not have an error but got one")
	}
}

func TestIsEmail(t *testing.T){
	postedData := url.Values{}
	form := New(postedData)
	//empty postedData
	form.IsEmail("x")
	if form.Valid(){
		t.Error("form shows valid email for non-existent field")
	}
	//wrong postedData
	postedData = url.Values{}
	postedData.Add("email", "xxxxx")
	form = New(postedData)

	form.IsEmail("email")
	if form.Valid(){
		t.Error("got an invalid email when we shouldn't have")
	}
	//correct postedData
	postedData = url.Values{}
	postedData.Add("email", "me@here.com")
	form = New(postedData)

	form.IsEmail("email")
	if !form.Valid(){
		t.Error("got an invalid email when we shouldn't have")
	}
}