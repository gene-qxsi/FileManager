package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

func main() {
	start()
}

func start() {
	for {
		search(scan())
	}
}

func search(data string) {
	if strings.HasPrefix(data, "cd") {
		move(data)
	} else if data == ".." || data == "back" {
		back()
	} else if strings.HasPrefix(data, "cat") {
		cat(data)
	} else if strings.HasPrefix(data, "delete") {
		delete(data)
	} else if strings.HasPrefix(data, "create") {
		create(data)
	} else if strings.HasPrefix(data, "ls") {
		ls(data)
	} else if strings.HasPrefix(data, "rename") {
		rename(data)
	} else if strings.HasPrefix(data, "copy") {
		copy(data)
	} else if strings.HasPrefix(data, "inf") {
		inf(data)
	} else if data == "restart" {
		restart()
	} else if data == "close" {
		close()
	} else {
		colorPrinter("команда <%s> не поддерживается приложением\n", color.FgRed, data)
	}
}

// получение информации о файле
func inf(data string) {
	args := strings.Fields(data)
	if len(args) != 2 {
		colorPrinter("команда <inf> используется не верно\n", color.FgRed)
		return
	}

	file := args[1]

	info, err := os.Stat(file)
	if err != nil {
		if os.IsNotExist(err) {
			colorPrinter("файл %s не найден", color.FgRed, file)
		}
		colorPrinter("не удалось получить информацию о файле %s", color.FgRed, file)
		return
	}

	colorPrinter("имя: %s\nразмер: %dKB\nвремя последнего изменения: %v\nправа: %o\nтип: %v\n", color.FgHiWhite, info.Name(), info.Size(), info.ModTime(), info.Mode().Perm(), info.Mode().Type())
}

// копирование
func copy(data string) {
	basepath, _ := os.Getwd()
	args := strings.Fields(data)
	if len(args) != 3 {
		colorPrinter("команда <copy> используется не верно\n", color.FgRed)
		return
	}

	oldpath := args[1] // файл который нужно скопировать
	newpath := args[2] // конечная деректория

	filename := filepath.Base(oldpath)

	err := os.Chdir(newpath)

	if err != nil {
		err := os.MkdirAll(newpath, 0777) // создаю конечную директорию если её не существует
		if os.IsPermission(err) {
			colorPrinter("Недостаточно прав для выполнения операции. Попробуйте запустить программу от имени администратора.\n", color.FgYellow)
			return
		}
		err = os.Chdir(newpath)
		if err != nil {
			colorPrinter("не удалось создать директорию %s и перейти в неё", color.FgRed, newpath)
			return
		}
	}

	file, err := os.Create(filename)
	if err != nil {
		colorPrinter("не удалось создать файл в конечной директории %s", color.FgRed, newpath)
		return
	}

	bytedata, readerr := os.ReadFile(oldpath)
	if readerr != nil {
		colorPrinter("не удалось считать данные из файла %s", color.FgRed, oldpath)
		return
	}
	err = os.WriteFile(file.Name(), bytedata, 0777)
	if err != nil {
		colorPrinter("не удалось записать данные в файл %s", color.FgRed, file.Name())
		return
	}
	os.Chdir(basepath)
	colorPrinter("успешно\n", color.FgCyan)
}

// перемещение/переименовывание
func rename(data string) {
	args := strings.Fields(data)
	if len(args) != 3 {
		colorPrinter("команда <rename> исопльзуется не верно\n", color.FgRed)
		return
	}

	oldpath := args[1]
	newpath := args[2]

	_, err := os.Stat(oldpath)
	if err != nil {
		colorPrinter("ошибка при получени информации о %s, возможно его не существует\n", color.FgRed, oldpath)
		return
	}

	err = os.Rename(oldpath, newpath)
	if err != nil {
		colorPrinter("возникла ошибка {%s} при перемещении/переименовывании файла\n", color.FgRed, err.Error())
		if os.IsPermission(err) {
			colorPrinter("Недостаточно прав для выполнения операции. Попробуйте запустить программу от имени администратора.\n", color.FgYellow)
		} else {
			colorPrinter("Проверьте, используется ли файл/директория в другой программе или имеет ли он атрибуты только для чтения.\n", color.FgYellow)
		}
		return
	}
	colorPrinter("успешно\n", color.FgCyan)
}

// информация по текущей дериктории
func ls(data string) {
	args := strings.Fields(data)
	if len(args) != 1 {
		colorPrinter("команда <ls> используетя не верно\n", color.FgRed)
		return
	}
	wd, _ := os.Getwd()
	files, err := os.ReadDir(wd)
	if err != nil {
		colorPrinter("не удается получить информацию о текущей директории\n", color.FgRed)
		return
	}

	for i, v := range files {
		colorPrinter("%d. %s\n", color.FgHiWhite, i+1, v.Name())
	}

}

// создание
// for upd> добавить возможность перезаписи или отмены действия
func create(data string) {
	perm := "0644"
	args := strings.Fields(data)
	if len(args) != 3 && len(args) != 4 {
		colorPrinter("команда <create> используется не верно\n", color.FgRed)
		return
	}

	if len(args) == 4 {
		perm = args[3]
	}

	createType := args[1]
	path := args[2]

	if createType == "file" {
		file, err := os.Create(path)
		if err != nil {
			colorPrinter("возникла ошибка при создании файла, файл %s не создан\n", color.FgRed, path)
			return
		}
		file.Close()
	} else if createType == "dir" {
		perm, err := strconv.Atoi(perm)
		if err != nil {
			perm = 0644
		}
		err = os.MkdirAll(path, os.FileMode(perm))
		if err != nil {
			colorPrinter("не удалось создать директорию %s, ошибка не известна\n", color.FgRed, path)
			return
		}
	}
	colorPrinter("успешно\n", color.FgCyan)
}

// удаление
func delete(data string) {
	args := strings.Fields(data)
	if len(args) != 2 {
		colorPrinter("функция <delete> используется не верно\n", color.FgRed)
		return
	}

	file := args[1]

	wd, _ := os.Getwd()

	if !filepath.IsAbs(file) {
		file = filepath.Join(wd, file)
	}

	_, err := os.Stat(file)
	if err != nil {
		if os.IsNotExist(err) {
			colorPrinter("пути %s не существует\n", color.FgRed, file)
		}
		return
	}

	err = os.RemoveAll(file)
	if err != nil {
		fmt.Println(os.IsNotExist(err))
		colorPrinter("удалить %s не получилось, ошибка: %s\n", color.FgRed, file, err.Error())
		return
	}

	colorPrinter("успешно\n", color.FgCyan)
}

// чтение с консоли
func scan() string {
	currentDirPrinter()
	reader := bufio.NewReader(os.Stdin)
	command, err := reader.ReadString('\n')
	if err != nil {
		colorPrinter("ошбика ввода\n", color.FgRed)
		return "nil pointer"
	}
	command = strings.TrimSpace(command)
	return command
}

// открытие файла
func cat(data string) {
	args := strings.Fields(data)

	if len(args) != 2 {
		colorPrinter("команда <cat> введена не верно\n", color.FgRed)
		return
	}

	file := args[1]
	wd, _ := os.Getwd()

	if !filepath.IsAbs(file) {
		file = filepath.Join(wd, file)
	}

	inf, err := os.Stat(file)
	if err != nil {
		if os.IsNotExist(err) {
			colorPrinter("пути %s не существует\n", color.FgRed, file)
		}
		return
	}

	if inf.IsDir() {
		colorPrinter("команда <cat> не применяется дли директорий\n", color.FgRed)
		return
	}

	var cmd *exec.Cmd = nil
	ext := filepath.Ext(file)

	if ext == ".txt" || ext == ".md" {
		cmd = exec.Command("notepad", file)
	} else if ext == ".mp3" {
		cmd = exec.Command("cmd", "/C", "start", file)
	} else if ext == ".exe" {
		fmt.Println(file)
	}
	if cmd != nil {
		err = cmd.Run()
		if err != nil {
			if ext == ".mp3" {
				colorPrinter("ошибка при попытке открыть медиа плеер\n", color.FgRed)
			} else if ext == ".txt" {
				colorPrinter("ошибка при попытке открыть блокнот\n", color.FgRed)
			} else if ext == ".exe" {
				colorPrinter("ошибка при попытке запустить .exe файл", color.FgRed)
			}
		}
	} else {
		colorPrinter("файл типа %s не может быть обработан\n", color.FgRed, ext)
	}
}

// передживение
func move(data string) {
	args := strings.Fields(data)
	if len(args) != 2 {
		colorPrinter("команда <cd> введена не верно\n", color.FgRed)
		return
	}

	newpath := args[1]
	wd, _ := os.Getwd()

	if !filepath.IsAbs(newpath) {
		newpath = filepath.Join(wd, newpath)
	}

	info, err := os.Stat(newpath)
	if err != nil {
		if os.IsNotExist(err) {
			colorPrinter("пути %s не существует\n", color.FgRed, newpath)
		}
		return
	}

	if !info.IsDir() {
		colorPrinter("%s не является дирректорией, переход не возможен\n", color.FgRed, info.Name())
		return
	}

	os.Chdir(newpath)

}

func back() {
	wd, _ := os.Getwd()
	os.Chdir(filepath.Dir(wd))
}

// прочее
func colorPrinter(text string, ccolor color.Attribute, args ...interface{}) {
	color.New(ccolor).Printf(text, args...)
}

func currentDirPrinter() {
	path, _ := os.Getwd()
	colorPrinter(path+">", color.FgGreen)
}

func restart() {
	path, err := os.Executable()
	if err != nil {
		colorPrinter("не удалось получить путь к текущему приложению\n", color.FgRed)
		return
	}

	cmd := exec.Command(path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	err = cmd.Start()
	if err != nil {
		colorPrinter("не удалось перезапустить приложение\n", color.FgRed)
		return
	}
	colorPrinter("приложение перезапускается\n", color.FgHiWhite)
	os.Exit(0)
}

func close() {
	colorPrinter("программа будет закрыта", color.FgHiWhite)
	os.Exit(0)
}
