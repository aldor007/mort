package middleware

import (
	"bytes"
	"github.com/aldor007/go-aws-auth"
	"github.com/aldor007/mort/pkg/config"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

type nextHandler struct {
	called bool
}

func (n *nextHandler) ServeHTTP(_ http.ResponseWriter, _ *http.Request) {
	n.called = true
}

func TestS3Auth_HandlerNoAuth(t *testing.T) {
	configData := config.GetInstance()
	configData.Load("./config.yml")

	s3 := NewS3AuthMiddleware(configData)

	next := nextHandler{}
	fn := s3.Handler(&next)

	req, _ := http.NewRequest("GET", "http://mort/local/test.jpg", nil)
	recorder := httptest.NewRecorder()
	fn.ServeHTTP(recorder, req)

	assert.True(t, next.called)
}

func TestS3Auth_Handler401(t *testing.T) {
	configData := config.GetInstance()
	configData.Load("./config.yml")

	s3 := NewS3AuthMiddleware(configData)

	next := nextHandler{}
	fn := s3.Handler(&next)

	req, _ := http.NewRequest("PUT", "http://mort/local/test.jpg", nil)
	recorder := httptest.NewRecorder()
	fn.ServeHTTP(recorder, req)

	assert.False(t, next.called)
	assert.Equal(t, recorder.Code, 401)
}

func TestS3Auth_Handler200S3(t *testing.T) {
	configData := config.GetInstance()
	configData.Load("./config.yml")

	s3 := NewS3AuthMiddleware(configData)

	next := nextHandler{}
	fn := s3.Handler(&next)

	req, _ := http.NewRequest("GET", "http://mort/local/test.jpg", nil)
	awsauth.SignS3(req, awsauth.Credentials{AccessKeyID: "acc", SecretAccessKey: "sec"})

	recorder := httptest.NewRecorder()
	fn.ServeHTTP(recorder, req)

	assert.True(t, next.called)
	assert.Equal(t, recorder.Code, 200)
}

func TestS3Auth_Handler200S3Put(t *testing.T) {
	configData := config.GetInstance()
	configData.Load("./config.yml")

	s3 := NewS3AuthMiddleware(configData)

	next := nextHandler{}
	fn := s3.Handler(&next)

	buf := bytes.Buffer{}
	buf.WriteString("aaaa-s3")

	req, _ := http.NewRequest("PUT", "http://mort/local/test.jpg", &buf)
	req.Header.Add("content-type", "image/jpg")
	awsauth.SignS3(req, awsauth.Credentials{AccessKeyID: "acc", SecretAccessKey: "sec"})

	recorder := httptest.NewRecorder()
	fn.ServeHTTP(recorder, req)

	assert.True(t, next.called)
	assert.Equal(t, recorder.Code, 200)
}

func TestS3Auth_Handler403S3Put(t *testing.T) {
	configData := config.GetInstance()
	configData.Load("./config.yml")

	s3 := NewS3AuthMiddleware(configData)

	next := nextHandler{}
	fn := s3.Handler(&next)

	buf := bytes.Buffer{}
	buf.WriteString("aaaa-s3")

	req, _ := http.NewRequest("PUT", "http://mort/local/test.jpg", &buf)
	awsauth.SignS3(req, awsauth.Credentials{AccessKeyID: "acc", SecretAccessKey: "sedc"})

	recorder := httptest.NewRecorder()
	fn.ServeHTTP(recorder, req)

	assert.False(t, next.called)
	assert.Equal(t, recorder.Code, 403)
}

func TestS3Auth_Handler401S3Put_2(t *testing.T) {
	configData := config.GetInstance()
	configData.Load("./config.yml")

	s3 := NewS3AuthMiddleware(configData)

	next := nextHandler{}
	fn := s3.Handler(&next)

	buf := bytes.Buffer{}
	buf.WriteString("aaaa-s3")

	req, _ := http.NewRequest("PUT", "http://mort/local/test.jpg", &buf)
	awsauth.SignS3(req, awsauth.Credentials{AccessKeyID: "acc2", SecretAccessKey: "sec"})

	recorder := httptest.NewRecorder()
	fn.ServeHTTP(recorder, req)

	assert.False(t, next.called)
	assert.Equal(t, recorder.Code, 401)
}

func TestS3Auth_HandlerS3LitBucket(t *testing.T) {
	configData := config.GetInstance()
	configData.Load("./config.yml")

	s3 := NewS3AuthMiddleware(configData)

	next := nextHandler{}
	fn := s3.Handler(&next)

	buf := bytes.Buffer{}
	buf.WriteString("aaaa-s3")

	req, _ := http.NewRequest("GET", "http://mort/", &buf)
	awsauth.SignS3(req, awsauth.Credentials{AccessKeyID: "acc", SecretAccessKey: "sec"})

	recorder := httptest.NewRecorder()
	fn.ServeHTTP(recorder, req)

	assert.False(t, next.called)
	assert.Equal(t, recorder.Code, 200)
	assert.Equal(t, recorder.HeaderMap.Get("content-type"), "application/xml")
}

func TestS3Auth_Handler200S3Put_v4(t *testing.T) {
	configData := config.GetInstance()
	configData.Load("./config.yml")

	s3 := NewS3AuthMiddleware(configData)

	next := nextHandler{}
	fn := s3.Handler(&next)

	buf := bytes.Buffer{}
	buf.WriteString("aaaa-s3")

	req, _ := http.NewRequest("PUT", "http://mort/local/test.jpg", &buf)
	awsauth.Sign4ForRegion(req, "mort", "s3", []string{}, awsauth.Credentials{AccessKeyID: "acc", SecretAccessKey: "sec"})

	recorder := httptest.NewRecorder()
	fn.ServeHTTP(recorder, req)

	assert.True(t, next.called)
	assert.Equal(t, recorder.Code, 200)
}

func TestS3Auth_Handler401S3Put_v4Query(t *testing.T) {
	configData := config.GetInstance()
	configData.Load("./config.yml")

	s3 := NewS3AuthMiddleware(configData)

	next := nextHandler{}
	fn := s3.Handler(&next)

	buf := bytes.Buffer{}
	buf.WriteString("aaaa-s3")

	req, _ := http.NewRequest("PUT", "http://mort/local/test.jpg", &buf)
	awsauth.PreSign(req, "mort", "s3", []string{}, awsauth.Credentials{AccessKeyID: "acc", SecretAccessKey: "sec"})

	recorder := httptest.NewRecorder()
	fn.ServeHTTP(recorder, req)

	assert.False(t, next.called)
	assert.Equal(t, recorder.Code, 401)
}

func TestS3Auth_Handler200S3Get_v4Query(t *testing.T) {
	configData := config.GetInstance()
	configData.Load("./config.yml")

	s3 := NewS3AuthMiddleware(configData)

	next := nextHandler{}
	fn := s3.Handler(&next)

	buf := bytes.Buffer{}
	buf.WriteString("aaaa-s3")

	req, _ := http.NewRequest("GET", "http://mort/local/test.jpg?X-Amz-Date=88888888888&X-Amz-Credential=acc/AWS", &buf)
	awsauth.PreSign(req, "mort", "s3", []string{}, awsauth.Credentials{AccessKeyID: "acc", SecretAccessKey: "sec"})


	recorder := httptest.NewRecorder()
	fn.ServeHTTP(recorder, req)

	assert.True(t, next.called)
	assert.Equal(t, recorder.Code, 200)
}

func TestS3Auth_Handler401S3Get_v4Query2(t *testing.T) {
	configData := config.GetInstance()
	configData.Load("./config.yml")

	s3 := NewS3AuthMiddleware(configData)

	next := nextHandler{}
	fn := s3.Handler(&next)

	buf := bytes.Buffer{}
	buf.WriteString("aaaa-s3")

	req, _ := http.NewRequest("GET", "http://mort/local/test.jpg?X-Amz-Date=88888888888&X-Amz-Credential=ac5c/AWS", &buf)
	awsauth.PreSign(req, "mort", "s3", []string{}, awsauth.Credentials{AccessKeyID: "acc", SecretAccessKey: "sec"})


	recorder := httptest.NewRecorder()
	fn.ServeHTTP(recorder, req)

	assert.False(t, next.called)
	assert.Equal(t, recorder.Code, 401)
}
