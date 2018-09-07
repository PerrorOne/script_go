package main

import (
	"archive/zip"
	"os"
	"strings"
	"io"
	"path/filepath"
	"os/exec"
	"flag"
	"net/http"
	"io/ioutil"
	"log"
	)

var (
	upxLinuxUrl = "http://github.com/upx/upx/releases/download/v3.95/upx-3.95-win64.zip"
)

type Info struct {
	FilePath   *string
	Output *string
	UpxUrl     string
	UpxName    string
	ExeFile    string
	HomePath   string
	upxPath    string
}

func unzip(archive, target string) error {
	reader, err := zip.OpenReader(filepath.Join(target, archive))
	if err != nil {
		log.Println(err, archive)
		return err
	}

	for _, file := range reader.File {
		path := filepath.Join(target, file.Name)
		if !strings.HasSuffix(path, "upx") && !strings.HasSuffix(path, "upx.exe") {
			continue
		}
		path = filepath.Join(target, "upx")

		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}

		dir := filepath.Dir(path)
		if len(dir) > 0 {
			if _, err = os.Stat(dir); os.IsNotExist(err) {
				err = os.MkdirAll(dir, 0755)
				if err != nil {
					return err
				}
			}
		}

		//---------------------end

		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		defer fileReader.Close()

		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		defer targetFile.Close()

		if _, err := io.Copy(targetFile, fileReader); err != nil {
			return err
		}
	}

	return nil
}

func homeWindows() string {
	drive := os.Getenv("HOMEDRIVE")
	path := os.Getenv("HOMEPATH")
	home := drive + path
	if drive == "" || path == "" {
		home = os.Getenv("USERPROFILE")
	}
	if home == "" {
		panic("HOMEDRIVE, HOMEPATH, and USERPROFILE are blank")
	}
	return home
}

func main() {
	// unzip
	info := &Info{}
	log.SetPrefix("[build] ")
	info.Output = flag.String("o", "", "需要编译后文件输出路径")
	info.FilePath = flag.String("f", "", "需要编译的文件路径")
	info.HomePath = homeWindows()
	info.upxPath = filepath.Join(info.HomePath, "upx")
	flag.Parse()
	_, err := os.Stat(filepath.Join(info.upxPath, "upx.exe"))
	if err == nil {
		// 存在直接调用
		info.upxPath = filepath.Join(info.upxPath, "upx.exe")
	} else {
		log.Println("No UPX found in the current environment,start downloading")
		err = os.MkdirAll(info.upxPath, 0777)
		if strings.HasSuffix(upxLinuxUrl, "zip") {
			resp, err := http.Get(upxLinuxUrl)
			log.Println("Download completed,unzipped")
			if err != nil {
				log.Println(err)
				return
			}
			f, err := os.Create(filepath.Join(info.upxPath, "upx.zip"))
			if err != nil{
				log.Println(err)
				return
			}
			d, err := ioutil.ReadAll(resp.Body)
			if err != nil{
				log.Println(err)
				return
			}
			f.Write([]byte(d))
			f.Close()
			unzip("upx.zip", info.upxPath)
			err = os.RemoveAll(filepath.Join(info.upxPath, "upx.zip"))
			os.Rename(filepath.Join(info.upxPath, "upx"), filepath.Join(info.upxPath, "upx.exe"))
			info.upxPath = filepath.Join(info.upxPath, "upx.exe")
			log.Println("unzippend complete, UPX file path: ", info.upxPath)
		}
	}
	log.Println("starting Compiled")
	old_path, file := filepath.Split(*info.FilePath)
	file = strings.Replace(file, ".go", "", 1)
	args := []string{"build", "-ldflags", `-s -w`}
	if *info.Output != ""{
		path, fileName := filepath.Split(*info.Output)
		if fileName == ""{
			fileName = file
		}
		output := filepath.Join(path, fileName)
		info.Output = &output
		args = append(args, output)
	}else{
		p := filepath.Join(old_path, file)
		info.Output = &p
		args = append(args, "-o")
		args = append(args, *info.Output)
		args = append(args, *info.FilePath)
	}
	execmd := exec.Command("go", args...)
	var d []string
	Environ := os.Environ()
	d = append(d, Environ...)
	d = append(d, "GOHOSTOS=linux")
	d = append(d, "CGO_ENABLED=0")
	d = append(d, "GOARCH=amd64")
	execmd.Env = d
	err = execmd.Run()
	if err != nil {
		log.Println(err)
		return
	}
	f, err := os.Stat(*info.Output)
	if err == nil{
		filesize := float64(f.Size()) / 1024.0 / 1024.0
		log.Printf("Compile completion,Compiled file size: %.2fM,Starting UPX compress...", filesize)
	}else{
		log.Println("build error: ", err)
		return
	}
	// 压缩
	cmd := exec.Command(info.upxPath, *info.Output)
	cmd.Env = Environ
	err = cmd.Run()
	if err != nil {
		log.Println(err)
		return
	}

	f, err = os.Stat(*info.Output)
	if err == nil{
		fileSize := float64(f.Size())  / 1024.0 / 1024.0
		log.Printf("Compress completion,Size of compressed file: %.2fM, File path:%s", fileSize, *info.Output)
	}else{
		log.Println("Compress error: ", err)
		return
	}
}
