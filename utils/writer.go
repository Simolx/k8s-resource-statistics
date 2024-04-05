package utils

import (
	"encoding/csv"
	"encoding/json"
	"example.com/dev/k8s/controllers"
	"fmt"
	"github.com/xuri/excelize/v2"
	"os"
	"path/filepath"
	"strconv"
)

const (
	filePerm      = 0600
	directoryPerm = 0700
)

func checkAndCreateDirectory(filePath string, isFile bool) error {
	fileDirectory := filePath
	if isFile {
		fileDirectory = filepath.Dir(fileDirectory)
	}
	if fileInfo, err := os.Stat(fileDirectory); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else if !fileInfo.IsDir() {
		return fmt.Errorf("path %q exists, but not directory", fileDirectory)
	}
	return os.MkdirAll(fileDirectory, directoryPerm)
}

func WriteJsonFile(content interface{}, filePath string) error {
	if err := checkAndCreateDirectory(filePath, true); err != nil {
		return err
	} else if contentJson, err := json.Marshal(content); err != nil {
		return err
	} else {
		return os.WriteFile(filePath, contentJson, filePerm)
	}
}

func WriteCsvFile(content [][]string, header []string, filePath string) error {
	if err := checkAndCreateDirectory(filePath, true); err != nil {
		return err
	} else if file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, filePerm); err != nil {
		return err
	} else {
		defer file.Close()
		csvWriter := csv.NewWriter(file)
		defer csvWriter.Flush()
		if len(header) > 0 {
			csvWriter.Write(header)
		}
		csvWriter.WriteAll(content)
	}
	return nil
}

func generateControllerInfo(controllerItem controllers.ControllerItem) []string {
	return []string{
		controllerItem.Namespace, controllerItem.ControllerType, controllerItem.Controller, strconv.Itoa(int(controllerItem.Replicas)),
		strconv.FormatInt(controllerItem.EmptyDir, 10), strconv.Itoa(controllerItem.Storage), strconv.FormatBool(controllerItem.StorageNoSize),
	}
}

func generateContainerInfo(controllerItem controllers.ControllerItem) [][]string {
	var result [][]string
	containerType := "initContainer"
	for _, container := range controllerItem.InitContainer {
		result = append(result, []string{
			containerType, container.Name, strconv.FormatInt(container.RequestCPU, 10), strconv.FormatInt(container.RequestMem, 10), strconv.FormatInt(container.RequestEphemeralStorate, 10),
			strconv.FormatInt(container.LimitCPU, 10), strconv.FormatInt(container.LimitMem, 10), strconv.FormatInt(container.LimitEphemeralStorate, 10),
		})

	}
	containerType = "container"
	for _, container := range controllerItem.Container {
		result = append(result, []string{
			containerType, container.Name, strconv.FormatInt(container.RequestCPU, 10), strconv.FormatInt(container.RequestMem, 10), strconv.FormatInt(container.RequestEphemeralStorate, 10),
			strconv.FormatInt(container.LimitCPU, 10), strconv.FormatInt(container.LimitMem, 10), strconv.FormatInt(container.LimitEphemeralStorate, 10),
		})
	}
	return result
}

func WriteExcelFile(content []controllers.ControllerItem, filePath string, sheet string) error {
	headers := []string{"namespace", "controllerType", "controller", "replicas", "emptyDir(m)", "storage(m)", "storageNoSize",
		"containerType", "containerName", "requestCpu", "requestMem(m)", "requestEphemeralStorage(m)", "limitCpu", "limitMem(m)", "limitEphemeralStorage(m)"}
	if err := checkAndCreateDirectory(filePath, true); err != nil {
		return err
	}
	excelFile := excelize.NewFile()
	defer excelFile.Close()
	if index, err := excelFile.NewSheet(sheet); err != nil {
		return err
	} else {
		excelFile.SetActiveSheet(index)
	}
	rowIndex := 1
	//	writer header
	for index, header := range headers {
		if cell, err := excelize.CoordinatesToCellName(index+1, rowIndex); err != nil {
			return err
		} else if err := excelFile.SetCellValue(sheet, cell, header); err != nil {
			return err
		}
	}
	rowIndex++
	for _, controllerItem := range content {
		columnIndex := 1
		records := len(controllerItem.Container) + len(controllerItem.InitContainer) - 1
		for _, controllerInfo := range generateControllerInfo(controllerItem) {
			if cell, err := excelize.CoordinatesToCellName(columnIndex, rowIndex); err != nil {
				return err
			} else if endCell, err := excelize.CoordinatesToCellName(columnIndex, rowIndex+records); err != nil {
				return err
			} else if err = excelFile.MergeCell(sheet, cell, endCell); err != nil {
				return err
			} else if err = excelFile.SetCellValue(sheet, cell, controllerInfo); err != nil {
				return err
			}
			columnIndex++
		}
		for _, record := range generateContainerInfo(controllerItem) {
			for recordColumn, column := range record {
				if cell, err := excelize.CoordinatesToCellName(columnIndex+recordColumn, rowIndex); err != nil {
					return err
				} else if err = excelFile.SetCellValue(sheet, cell, column); err != nil {
					return err
				}
			}
			rowIndex++
		}
	}
	if err := excelFile.SaveAs(filePath); err != nil {
		return err
	}
	return nil
}
