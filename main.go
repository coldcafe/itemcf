package main

import (
	"rec/itemcf"
)

func main() {
	c := itemcf.NewItemCF()
	c.GetDataset("../ml-latest-small/ratings.csv")
	c.CalcMovieSim()
	c.Evaluate()
}
