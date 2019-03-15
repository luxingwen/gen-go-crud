package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

type Model struct {
	Name   string
	Fields []*Field
}

type Field struct {
	Name string
	Type string
	IsPk bool
	From string
}

type GenModel struct {
}

func (g *GenModel) GenModel(filename string) (rm []*Model, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return
	}
	buf := bufio.NewReader(f)
	var strs []string
	for {
		line, err := buf.ReadString('\n')
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") {
			continue
		}
		strs = append(strs, strings.TrimSpace(line))
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return rm, err
			}
		}
	}
	fmt.Println(strs)
	rm = g.getModels(strs)
	return
}

func (g *GenModel) getModels(strs []string) (rm []*Model) {
	var start int
	for i, item := range strs {
		if strings.HasPrefix(item, "type") && strings.HasSuffix(item, "{") {
			start = i
		}
		if item == "}" {
			rm = append(rm, g.getModel(strs[start:i]))
		}

	}
	return
}

var mType map[string]string = map[string]string{
	"int":    "int",
	"string": "string",
}

func (g *GenModel) getModel(strs []string) (m *Model) {
	m = new(Model)
	for _, item := range strs {
		if strings.HasPrefix(item, "type") && strings.HasSuffix(item, "{") {
			m.Name = strings.TrimRight(strings.TrimLeft(item, "type"), "struct{")
			m.Name = strings.TrimSpace(m.Name)
			continue
		}
		m.Fields = append(m.Fields, g.getField(item))
	}
	return
}

func (g *GenModel) getField(line string) (f *Field) {
	f = new(Field)
	for _, item := range strings.Fields(line) {
		if v, ok := mType[item]; ok {
			f.Type = v
			continue
		}
		if strings.HasPrefix(item, "pk") {
			f.IsPk = true
			continue
		}
		if strings.HasPrefix(item, "from") {
			f.From = strings.Split(item, ":")[1]
			continue
		}
		f.Name = item
	}
	return
}

func (m *Model) genCurd() (err error) {
	str := "package models\n\ntype " + m.Name + " struct {"
	for _, itemF := range m.Fields {
		str += fmt.Sprintf("\n\t%s\t%s\t`gorm:\"cloumn:%s\"`", strings.Title(itemF.Name), itemF.Type, itemF.Name)
	}
	str += "\n}\n"

	str += "func Get" + m.Name + "() *" + m.Name + " {\nreturn &" + m.Name + "{}\n}\n\n"
	str += fmt.Sprintf("func (m *%s)TableName()string {\n\treturn \"%s\"\n}", m.Name, strings.ToLower(m.Name))
	str += fmt.Sprintf("\nfunc (m *%s)GetOne(id int64) (r *%s, err error){\n", m.Name, m.Name)
	str += "\terr=db.Table(m.TableName()).Where(\"id = ?\", id).First(&r).Error\n\treturn\n}\n"

	str += fmt.Sprintf("\nfunc (m *%s)Update(id int64, m map[string]interface{}) (err error){\n", m.Name)
	str += "\terr=db.Table(m.TableName()).Where(\"id = ?\", id).Update(m).Error\n\treturn\n}\n"

	f, err := os.OpenFile("models/"+strings.ToLower(m.Name)+".go", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return
	}
	f.WriteString(str)
	f.Close()
	return
}

func main() {
	g := new(GenModel)
	r, err := g.GenModel("models.conf")
	if err != nil {
		log.Fatal(err)
	}
	for _, item := range r {
		item.genCurd()
		// fmt.Println(item)
		// for _, itemv := range item.Fields {
		// 	fmt.Println(itemv)
		// }
	}
	//r[0].genCurd()
}
