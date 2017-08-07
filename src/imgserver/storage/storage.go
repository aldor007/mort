package storage

import (
	"fmt"
	"os"
	"regexp"
	"io/ioutil"
	"mime"
	"path"

	"imgserver/object"
	"imgserver/response"
	"imgserver/config"
)

var isUrl_RE = regexp.MustCompile("http://")
const notFound = "{\"error\":\"not found\"}"
const internalError = "{\"error\":\"internal error\"}"


func Get(obj *object.FileObject) (*response.Response) {
	key := obj.Key
	fmt.Printf("GET %s sc", key)
	if isUrl_RE.MatchString(key) {
		return response.New(400, nil, fmt.Errorf("Not implemented"))
	}

	data, err := getFromDisk(key)
	if os.IsNotExist(err) {
		fmt.Println(err)
		return response.New( 404, []byte(notFound), nil)
	} else if err != nil {
		return response.New(503, []byte(internalError), err)
	}

	return prepareResponse(obj, data)
}

func getFromDisk(filePath string) ([]byte, error) {
	fmt.Println("AAAAAA "+ config.GetInstance().LocalFilesPath +"AAAA")
	return ioutil.ReadFile(config.GetInstance().LocalFilesPath + filePath)
}

func prepareResponse(obj *object.FileObject, data []byte) (*response.Response) {
	res := response.New(200, data, nil)
	res.SetContentType(mime.TypeByExtension(path.Ext(obj.Key)))
	return res
}