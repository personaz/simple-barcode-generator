package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/fogleman/gg"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/oned"
	"golang.org/x/image/font/inconsolata"
)

type Apps struct {
	model  *Model
	app    fyne.App
	window fyne.Window
}

func (apps *Apps) getPath() string {
	// exec, err := os.Executable()
	exec, err := os.Getwd()
	if err != nil {
		panic("Gagal jalankan Applikasi")
	}
	return exec
}

func (apps *Apps) getJsonPath() string {
	exec := apps.getPath()
	return filepath.Join(exec, "data.json")
}

func (apps *Apps) initApp() {
	dataJson := apps.getJsonPath()
	if _, err := os.Stat(dataJson); os.IsNotExist(err) {
		panic("Data json tidak ditemukan")
	}

	fileContent, err := os.ReadFile(dataJson)
	if err != nil {
		panic("Gagal membaca file data.json")
	}
	err = json.Unmarshal(fileContent, &apps.model)
	if err != nil {
		panic("Gagal konversi data.json")
	}
}

func (apps *Apps) getProductName() []string {
	products := []string{}
	for _, row := range apps.model.Items {
		products = append(products, row.Name)
	}
	return products
}

func (apps *Apps) getProduct(index int) Item {
	return apps.model.Items[index]
}

func (apps *Apps) Start() {
	apps.app = app.NewWithID("barcode-generator")
	apps.window = apps.app.NewWindow("Barcode Generator")
	entryFrom := widget.NewEntry()
	entryFrom.Disabled()
	entryUntil := widget.NewEntry()
	selCode := widget.NewSelect(apps.getProductName(), nil)
	selCode.OnChanged = func(s string) {
		selItem := apps.getProduct(selCode.SelectedIndex())
		entryFrom.SetText(fmt.Sprintf("%d", selItem.LastBarcode+1))
	}
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Nama Barang", Widget: selCode},
			{Text: "Barcode Dari", Widget: entryFrom},
			{Text: "Barcode Hingga", Widget: entryUntil},
		},
		OnSubmit: func() {
			saveDialog := dialog.NewFolderOpen(func(lu fyne.ListableURI, err error) {
				if err == nil && lu != nil {
					savePath := lu.Path()
					item := apps.getProduct(selCode.SelectedIndex())
					from := item.LastBarcode + 1
					until, err := strconv.Atoi(entryUntil.Text)
					if err != nil {
						panic("Barcode Hingga tidak diketahui")
					}
					for i := from; i <= until; i++ {
						code := fmt.Sprintf(item.Format, i)
						apps.generateImage(item, code, savePath)
					}
					apps.model.Items[selCode.SelectedIndex()].LastBarcode = until
					apps.saveModelJson()
				}
			}, apps.window)
			saveDialog.Show()
		},
	}
	apps.window.SetContent(form)
	apps.window.ShowAndRun()
}

func (apps *Apps) saveModelJson() {
	json, err := json.MarshalIndent(apps.model, "", "  ")
	if err != nil {
		fmt.Println(err.Error())
	}
	err = os.WriteFile(apps.getJsonPath(), json, 0644)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func (apps *Apps) generateBarcode(code string) *gozxing.BitMatrix {
	writer := oned.NewCode128Writer()
	barcode, err := writer.Encode(code, gozxing.BarcodeFormat_CODE_128, 250, 50, nil)
	if err != nil {
		panic("Gagal membuat barcode")
	}
	return barcode
}

func (apps *Apps) generateImage(item Item, code string, savePath string) {
	width, height := 300, 160
	barcode := apps.generateBarcode(code)
	dc := gg.NewContext(width, height)
	dc.SetRGB(1, 1, 1)
	dc.Clear()
	dc.SetRGB(0, 0, 0)
	dc.SetFontFace(inconsolata.Bold8x16)
	dc.DrawStringAnchored("Type", float64(width/2), 16, 0.5, 0)
	dc.SetFontFace(inconsolata.Bold8x16)
	dc.DrawStringAnchored(item.Code, float64(width/2), 34, 0.5, 0)
	dc.DrawImageAnchored(barcode, width/2, barcode.GetHeight(), 0.5, 0)
	dc.DrawStringAnchored(code, float64(width/2), float64(barcode.GetHeight()+46), 0.5, 1)
	filename := fmt.Sprintf("%s-%s.jpg", item.Code, code)
	saveDir := filepath.Join(savePath, filename)
	dc.Clip()
	dc.SavePNG(saveDir)
}

func NewApps() Apps {
	app := Apps{}
	app.initApp()
	return app
}
