package main

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"github.com/urfave/cli/v2"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	GOLANGPROXYURL = "https://goproxy.cn"
	MODNAME        = "github.com/xml520/wqadmin"
	ProjectName    = "wqadmin"
)

func main() {
	app := &cli.App{
		Name:     "wqadmin",
		Usage:    "后台管理系统",
		HideHelp: true,
		Commands: []*cli.Command{
			newInit(),
		},
	}
	app.Run(os.Args)
}
func newInit() *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "初始化新项目",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "name",
				Usage:    "项目名称",
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			vers, err := getWqadminVersion()
			if err != nil {
				return errors.New("无法获取版本信息 " + err.Error())
			}
			fmt.Println("--wqadmin version：" + vers[len(vers)-1])
			zipReader, err := getWqAdminZip(vers[len(vers)-1])
			if err != nil {
				return errors.New("无法下载资源文件 " + err.Error())
			}
			err = export(c.String("name"), zipReader.File)
			if err != nil {
				return errors.New("无法导出文件 " + err.Error())
			}
			fmt.Println("--wqadmin download success")
			return nil
		},
	}
}
func getWqAdminZip(ver string) (*zip.Reader, error) {
	url := GOLANGPROXYURL + "/" + MODNAME + "/@v/" + ver + ".zip"
	get, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer get.Body.Close()
	var buf bytes.Buffer
	if _, err = io.Copy(&buf, get.Body); err != nil {
		return nil, err
	}
	return zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))

}
func getWqadminVersion() ([]string, error) {
	url := GOLANGPROXYURL + "/" + MODNAME + "/@v/list"
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	resText, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return strings.Split(string(resText), "\n"), nil
}
func export(name string, files []*zip.File) error {

	for _, file := range files {

		if file.FileInfo().IsDir() {
			continue
		}
		if strings.Index(file.Name, "wqadmincli") != -1 {
			continue
		}
		outFilename := filepath.Join("./"+name, file.Name)
		switch {
		case strings.HasSuffix(file.Name, ".go"), strings.HasSuffix(file.Name, "go.mod"):
			r, err := replaceGoFile(name, file)
			if err != nil {
				return err
			}
			if err = writerFile(outFilename, r); err != nil {
				return err
			}
		default:
			zr, err := file.Open()
			if err != nil {
				return err
			}
			defer zr.Close()
			if err = writerFile(outFilename, zr); err != nil {
				return err
			}
		}
	}
	return nil
}
func writerFile(filename string, reader io.Reader) error {
	os.MkdirAll(filepath.Dir(filename), 0755)
	outFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer outFile.Close()
	_, err = io.Copy(outFile, reader)
	return err
}
func replaceGoFile(modName string, file *zip.File) (io.Reader, error) {
	f, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	buf, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	buf = bytes.Replace(buf, []byte(ProjectName), []byte(modName), -1)
	return bytes.NewReader(buf), nil
}
