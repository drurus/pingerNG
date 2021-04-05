package drivefile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// func ReadFromFile(fname string) error {
func ReadFromFile(fname string, fn func(fname, tab string) error) error {
	dat, err := os.ReadFile(fname)
	if err != nil {
		return err
	}
	hosts_list := strings.Split(string(dat), "\n")
	for _, hsl := range hosts_list {
		// TODO добавление в Redis
		tab := filepath.Base(fname)
		if err = fn(hsl, tab); err != nil {
			return err
		}
	}
	return nil
}

func LoadDirectory(dirpath string, fn func(fname, tab string) error) error {
	err := filepath.Walk(dirpath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				// чтение файла
				// fmt.Println(path, info.Size(), s)
				fmt.Println("Reading file ", path)
				// tab := filepath.Dir(path)
				// fmt.Println("DIR ", tab)
				if err = ReadFromFile(path, fn); err != nil {
					return err
				}
			}
			return nil
		})
	if err != nil {
		return err
	}
	return nil
}
