package main

type Model struct {
	Items []Item `json:"items"`
}

type Item struct {
	Name        string `json:"nama_produk"`
	Code        string `json:"kode"`
	Format      string `json:"format"`
	LastBarcode int    `json:"last_barcode"`
}
