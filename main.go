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
	// бесконченый цикл для избежания использования рекурсии, упрощения
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
	} else if data == "ls" {
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
		colorPrinter("команда <inf> используется неверно\n", color.FgRed)
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

// копирование *файлов
func copy(data string) {
	args := strings.Fields(data)
	if len(args) != 3 {
		colorPrinter("команда <copy> используется неверно\n", color.FgRed)
		return
	}

	oldpath := args[1] // файл который нужно скопировать
	newpath := args[2] // конечная директория

	if _, err := os.Stat(oldpath); err != nil {
		colorPrinter("файл %s не существует\n", color.FgRed, oldpath)
		return
	}

	err := os.MkdirAll(newpath, 0777)
	if err != nil {
		if os.IsPermission(err) {
			colorPrinter("Недостаточно прав для выполнения операции. Запустите программу от имени администратора.\n", color.FgYellow)
		} else {
			colorPrinter("не удалось создать директорию %s\n", color.FgRed, newpath)
		}
		return
	}

	// Определяю имя конечного файла
	filename := filepath.Base(oldpath)
	destination := filepath.Join(newpath, filename)

	// считывание данных из исходного файла
	bytedata, readerr := os.ReadFile(oldpath)
	if readerr != nil {
		colorPrinter("не удалось считать данные из файла %s\n", color.FgRed, oldpath)
		return
	}

	// Запись данных в новый файл
	writeerr := os.WriteFile(destination, bytedata, 0666)
	if writeerr != nil {
		colorPrinter("не удалось записать данные в файл %s\n", color.FgRed, destination)
		return
	}

	colorPrinter("Файл успешно скопирован в %s\n", color.FgCyan, destination)
}

// перемещение/переименовывание
func rename(data string) {
	args := strings.Fields(data)
	if len(args) != 3 {
		colorPrinter("команда <rename> исопльзуется неверно\n", color.FgRed)
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
		colorPrinter("команда <ls> используетя неверно\n", color.FgRed)
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
		colorPrinter("команда <create> используется неверно\n", color.FgRed)
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
		colorPrinter("функция <delete> используется неверно\n", color.FgRed)
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
		colorPrinter("команда <cat> введена неверно\n", color.FgRed)
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
	ext := filepath.Ext(file)

	var cmd *exec.Cmd
	// я бы хотел переписать с использованием реестра винды
	// но пока не могу
	switch ext {
	case ".txt":
		cmd = exec.Command("notepad.exe", file)
	case ".mp3":
		cmd = exec.Command("wmplayer", file)
	case ".exe":
		cmd = exec.Command(file)
	default:
		colorPrinter("файл типа %s не может быть обработан\n", color.FgRed, ext)
		return
	}

	err = cmd.Run()
	if err != nil {
		colorPrinter("ошибка при попытке открыть файл %s: %v\n", color.FgRed, file, err)
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
