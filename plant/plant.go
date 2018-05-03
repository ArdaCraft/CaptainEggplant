package plant

import (
	"io/ioutil"
	"encoding/json"
	"fmt"
	"os"
)

type Plants struct {
	Main string          `json:"main"`
	All  map[string]bool `json:"all"`
}

func New() *Plants {
	var plants *Plants

	d, e := ioutil.ReadFile("plants.json")
	if e == nil {
		e = json.Unmarshal(d, plants)
		if e == nil {
			return plants
		}
	}

	return &Plants{Main: "", All: make(map[string]bool)}
}

func (p *Plants) Save() {
	d, e := json.MarshalIndent(p, "", "  ")
	if e != nil {
		fmt.Println(e)
		return
	}

	e = ioutil.WriteFile("plants.json", d, os.ModePerm)
	if e != nil {
		fmt.Println(e)
		return
	}
}